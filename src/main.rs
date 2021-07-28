use lorri::{
    builder, cas::ContentAddressable, nix::options::NixOptions, project::Project,
    watch::WatchPathBuf, AbsPathBuf, NixFile,
};
use std::any::Any;
use std::env;
use std::fs;
use std::iter::FromIterator;
use std::path::PathBuf;
use walkdir::WalkDir;

fn main() {
    let args: Vec<String> = env::args().collect();
    let tempdir = tempfile::tempdir().expect("tempfile::tempdir() failed us!");

    let project_file = &args[1];
    let project = project(
        project_file,
        &lorri::AbsPathBuf::new(tempdir.path().to_owned()).unwrap(),
    );
    let p = env::current_dir().unwrap().to_str().unwrap().to_string();

    let output = builder::run(&project.nix_file, &project.cas, &NixOptions::empty())
        .unwrap()
        .referenced_paths;
    let v = output
        .into_iter()
        .filter(|i| !i.as_ref().starts_with("/nix/store"))
        .collect::<Vec<_>>();
    for (_pos, e) in v.iter().enumerate() {
        let recursive_paths = match e {
            WatchPathBuf::Recursive(e) => {
                walk(&e.as_os_str().to_os_string().into_string().unwrap(), &p)
            }
            WatchPathBuf::Normal(e) => println!(
                "{}",
                strip(
                    &e.as_os_str()
                        .to_os_string()
                        .into_string()
                        .unwrap()
                        .strip_prefix(&p)
                        .unwrap()
                )
            ),
        };
    }
}

fn project(name: &str, cache_dir: &AbsPathBuf) -> Project {
    let test_root = AbsPathBuf::new(PathBuf::from_iter(&[env!("CARGO_MANIFEST_DIR"), name]))
        .expect("CARGO_MANIFEST_DIR was not absolute");
    let cas_dir = cache_dir.join("cas").to_owned();
    fs::create_dir_all(&cas_dir).expect("failed to create CAS directory");
    Project::new(
        NixFile::from(test_root),
        &cache_dir.join("gc_roots"),
        ContentAddressable::new(cas_dir).unwrap(),
    )
    .unwrap()
}

fn strip(s: &str) -> &str {
    let mut chars = s.chars();
    chars.next();
    chars.as_str()
}

fn walk(s: &str, p: &str) {
    for entry in WalkDir::new(s) {
        let entry = entry.unwrap();
        if entry.file_type().is_file() {
            println!(
                "{}",
                strip(
                    entry
                        .path()
                        .as_os_str()
                        .to_os_string()
                        .into_string()
                        .unwrap()
                        .strip_prefix(&p)
                        .unwrap()
                )
            );
        };
    }
}
