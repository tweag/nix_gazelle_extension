use std::env;
use std::path::{Path, PathBuf};

use lorri::watch::WatchPathBuf;
use lorri::AbsPathBuf;
use serde::Serialize;

#[derive(Serialize)]
struct DepSet {
  kind: String,
  files: Vec<String>,
}

#[derive(Serialize)]
struct DepSets {
  depsets: Vec<DepSet>,
}

#[derive(Debug, Clone)]
pub enum ErrorKind {
  AbsPathBufError,
  IoError,
  LorriError,
  NoneError,
  Other,
  SerializationError,
}

#[derive(Debug, Clone)]
pub struct ScanError {
  pub kind: ErrorKind,
  pub msg: String,
}

impl ScanError {
  pub fn new(kind: ErrorKind, msg: String) -> Self {
    Self { kind, msg }
  }

  pub fn new_from_str(kind: ErrorKind, msg: &str) -> Self {
    Self {
      kind,
      msg: msg.to_string(),
    }
  }
}

impl std::fmt::Display for ScanError {
  fn fmt(&self, f: &mut std::fmt::Formatter) -> std::fmt::Result {
    write!(f, "{}", self.msg)
  }
}

impl From<std::io::Error> for ScanError {
  fn from(error: std::io::Error) -> Self {
    ScanError::new(ErrorKind::IoError, error.to_string())
  }
}

impl From<lorri::error::BuildError> for ScanError {
  fn from(error: lorri::error::BuildError) -> Self {
    ScanError::new(ErrorKind::LorriError, error.to_string())
  }
}

impl From<serde_json::Error> for ScanError {
  fn from(error: serde_json::Error) -> Self {
    ScanError::new(ErrorKind::SerializationError, error.to_string())
  }
}

impl From<PathBuf> for ScanError {
  fn from(pathbuf: PathBuf) -> Self {
    ScanError::new(
      ErrorKind::AbsPathBufError,
      format!("Unable to create absolute path from: '{:?}'", pathbuf),
    )
  }
}

fn _prepare_tmpdir_root() -> Result<tempfile::TempDir, ScanError> {
  tempfile::Builder::new()
    .prefix("scan-nix")
    .tempdir()
    .map_or_else(|e| Err(ScanError::from(e)), |temp_dir| Ok(temp_dir))
}

fn _get_metadata(
  tmp_dir: &tempfile::TempDir,
  project_nix_file: &str,
) -> Result<lorri::project::Project, ScanError> {
  let nix_file_path = AbsPathBuf::new(PathBuf::from(project_nix_file))?;
  let cas_tmp_dir_path = AbsPathBuf::new(tmp_dir.path().clone().join("caas"))?;
  let gc_roots_tmp_dir_path =
    AbsPathBuf::new(tmp_dir.path().clone().join("gc_roots"))?;

  std::fs::create_dir(&cas_tmp_dir_path)?;
  std::fs::create_dir(&gc_roots_tmp_dir_path)?;

  let project = lorri::project::Project::new(
    lorri::NixFile::from(nix_file_path),
    &gc_roots_tmp_dir_path,
    lorri::cas::ContentAddressable::new(cas_tmp_dir_path)?,
  )?;
  Ok(project)
}

fn _get_all_paths_without_nix_store(
  project: &lorri::project::Project,
) -> Result<Vec<WatchPathBuf>, ScanError> {
  let build_result = lorri::builder::run(
    &project.nix_file,
    &project.cas,
    &lorri::nix::options::NixOptions::empty(),
  )?;
  Ok(
    build_result
      .referenced_paths
      .into_iter()
      .filter(|path| !path.as_ref().starts_with("/nix/store"))
      .collect::<Vec<WatchPathBuf>>(),
  )
}

fn _get_nix_project_path(
  project: &lorri::project::Project,
) -> Result<AbsPathBuf, ScanError> {
  match project.nix_file.as_absolute_path().to_path_buf().parent() {
    Some(parent_path) => Ok(AbsPathBuf::new(parent_path.to_path_buf())?),
    _ => Err(ScanError::new_from_str(
      ErrorKind::NoneError,
      "Value not found!",
    )),
  }
}

fn _is_child_of_nix_project(
  project: &lorri::project::Project,
  file_path: &AbsPathBuf,
) -> bool {
  match _get_nix_project_path(project) {
    Ok(project_path) => file_path
      .as_absolute_path()
      .starts_with(project_path.as_absolute_path()),
    _ => false,
  }
}

fn _to_list_of_direct_bzl_deps(
  project: &lorri::project::Project,
  file_paths: Vec<AbsPathBuf>,
) -> Result<Vec<String>, ScanError> {
  let parsed_file = &env::var("NIX_FILE").unwrap().to_owned();
  let nix_project_file = Path::new(parsed_file).to_path_buf();
  let nix_project_path = nix_project_file.parent().unwrap();

  Ok(
    file_paths
      .into_iter()
      .filter_map(|path| {
        match path.as_absolute_path().strip_prefix(nix_project_path) {
          Ok(pth) => match pth.to_str() {
            Some(p) => Some(String::from(p)),
            None => None,
          },
          _ => None,
        }
      })
      .collect::<Vec<String>>(),
  )
}

fn _get_workspace_root_path() -> Result<AbsPathBuf, ScanError> {
  // TODO: Fix this
  let cwd = std::env::current_dir()?;
  let path = AbsPathBuf::new(PathBuf::from(cwd))?;
  Ok(path)
}

fn _to_bazel_label(path: &std::path::Path, bazel_packages: &std::collections::HashSet<String>) -> Option<String> {
  match path.file_name() {
    Some(target_raw) => {
      let file = target_raw.to_str()?;
      let mut package: String = "".to_string();

      let mut path_ancestors = path.ancestors().into_iter().collect::<Vec<&Path>>();
      path_ancestors.reverse();

      for pth in path_ancestors.clone() {
        let sub_path = pth.to_str()?;
        if !bazel_packages.contains(sub_path) {
          continue;
        }

        package = String::from(sub_path);
        break;
      }

      let mut target: String = path_ancestors
        .last()?
        .to_str()?
        .replace(file, "")
        .replace(&package, "");

      match target.find("/") {
        Some(i) => {
          target.replace_range(i..(i+1), "");
        },
        None => (),
      }

      if package.is_empty() {
        Some(format!("//:{}", file))
      } else {
        Some(format!("//{}:{}{}", package, target, file))
      }
    }
    None => None,
  }
}

fn _to_list_of_bzl_deps(
  file_paths: Vec<AbsPathBuf>,
  bazel_packages: &std::collections::HashSet<String>
) -> Result<Vec<String>, ScanError> {
  let workspace_root = _get_workspace_root_path()?;
  let workspace_root_path = workspace_root.as_absolute_path();

  Ok(
    file_paths
      .into_iter()
      .filter_map(|path| {
        match path.as_absolute_path().strip_prefix(workspace_root_path) {
          Ok(pth) => _to_bazel_label(pth, bazel_packages),
          _ => None,
        }
      })
      .collect::<Vec<String>>(),
  )
}

fn _scan() -> Result<(), ScanError> {
  let args: Vec<String> = std::env::args().collect::<Vec<String>>();
  let project_file = args.get(1).ok_or(ScanError::new_from_str(
    ErrorKind::NoneError,
    "Value not found!",
  ))?;
  let tempdir = _prepare_tmpdir_root()?;

  let project = _get_metadata(&tempdir, project_file)?;

  let project_files = _get_all_paths_without_nix_store(&project)?
    .into_iter()
    .flat_map(|path| {
      match path {
        WatchPathBuf::Normal(pth) => vec![pth],
        // If path is a directory, translate it to list of files contained
        // within it
        WatchPathBuf::Recursive(pth) => walkdir::WalkDir::new(pth)
          .into_iter()
          .filter_map(|result| match result {
            Ok(entry) => {
              if !entry.file_type().is_dir() {
                Some(entry.into_path())
              } else {
                None
              }
            }
            _ => None,
          })
          .collect::<Vec<PathBuf>>(),
      }
    })
    .map(AbsPathBuf::new)
    .filter_map(|result| match result {
      Ok(pth) => Some(pth),
      _ => None,
    })
    .collect::<Vec<AbsPathBuf>>();

  let (project_children, mut project_deps): (Vec<AbsPathBuf>, Vec<AbsPathBuf>) =
    project_files
      .into_iter()
      .partition(|file_path| _is_child_of_nix_project(&project, file_path));
  // Every children is also a project dep
  project_deps.extend(project_children.clone());


  let workspace_root = _get_workspace_root_path()?;
  let workspace_root_path = workspace_root.as_absolute_path();
  let bazel_packages = project_deps.clone().into_iter().filter_map(|pth|{
    match pth.as_absolute_path().strip_prefix(workspace_root_path) {
      Ok(ppth) => { 
        match ppth.to_str() {
          Some(p) => {
            if p.contains("default.nix") {
              let package = String::from(p).replace("default.nix", "");
              if !package.is_empty() {
                Some(String::from(&package[..package.len()-1]))
              } else {
                None
              }
            } else {
              None
            }
          },
          _ => None,
        }
      },
      Err(_) => None,
    }
  }).collect::<std::collections::HashSet<String>>();

  // TODO: Better naming
  let all_deps = DepSets {
    depsets: vec![
      DepSet {
        kind: "recursive".to_string(),
        files: _to_list_of_bzl_deps(project_deps, &bazel_packages)?,
      },
      DepSet {
        kind: "direct".to_string(),
        files: _to_list_of_direct_bzl_deps(&project, project_children)?,
      },
    ],
  };

  let all_deps_repr = serde_json::to_string(&all_deps)?;
  println!("{}", all_deps_repr);

  Ok(())
}

fn main() {
  match _scan() {
    Ok(_) => (),
    Err(error) => println!("{:?}", error),
  }
}

#[cfg(test)]
mod tests {
  use super::*;
  use rstest::rstest;

  #[rstest]
  #[case("z", "//:z")]
  #[case("a/b", "//a:b")]
  #[case("c/d/e", "//c/d:e")]
  #[case("f/g/h/i", "//f/g/h:i")]
  #[case("j/k/l/m", "//j/k:l/m")]
  #[case("n/o/p/r/s/q", "//n/o/p:r/s/q")]
  fn test_to_bazel_label(
    #[case] input: String,
    #[case] expected_bzl_lbl: String,
  ) {
    let bazel_packages: std::collections::HashSet<String> = vec![
      "a".to_string(),
      "c/d".to_string(), 
      "f/g/h".to_string(),
      "j/k".to_string(),
      "n/o/p".to_string(),
    ].into_iter().collect();
    let input_path = std::path::Path::new(&input);

    match _to_bazel_label(input_path, &bazel_packages) {
      Some(bazel_label) => assert_eq!(expected_bzl_lbl, bazel_label),
      _ => assert!(false),
    }
  }
}
