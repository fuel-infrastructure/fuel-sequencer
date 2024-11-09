use std::{
    collections::HashMap,
    env, fs, io,
    path::{Path, PathBuf}, process::Command,
};

use regex::{Captures, Regex};

/// Module tree helper
#[derive(Default, Debug)]
struct Tree {
    /// This node of the tree has an associated module in this file
    module: Option<String>,
    /// Submodules of this module
    items: HashMap<String, Tree>,
}
impl Tree {
    /// Add a new file to the tree
    fn push(&mut self, path: &[&str], filename: String) {
        if path.is_empty() {
            self.module = Some(filename);
        } else {
            let key = path[0].to_owned();
            self.items
                .entry(key)
                .or_default()
                .push(&path[1..], filename);
        }
    }

    /// Combine all files in the tree into Rust source code.
    fn compile(&self) -> String {
        let mut result = String::new();

        // Recursively compile all submodules
        for (submodule, subtree) in self.items.iter() {
            result.push_str(&format!("pub mod {submodule} {{\n"));
            result.push_str(&subtree.compile());
            result.push_str("}\n");
        }

        // If this node has associated protobuf code, include it
        if let Some(module) = self.module.as_ref() {
            let src = read_and_preprocess(format!("protos/{module}")).unwrap();
            result.push_str(&src);
        }

        result
    }
}

fn read_and_preprocess<P: AsRef<Path>>(path: P) -> io::Result<String> {
    Command::new("rustfmt").args(["--edition", "2021", path.as_ref().to_str().unwrap()]).status()?;

    let src = fs::read(path).unwrap();
    let result = String::from_utf8_lossy(&src).into_owned();

    // Replace buf-generated includes with actual file content,
    // so everything is in one file that's easy to include.
    let re = Regex::new(r#"include!\("([^"]*)"\);"#).unwrap();

    let replacement = |caps: &Captures| -> String {
        let incl = caps[1].to_string();
        read_and_preprocess(format!("protos/{incl}")).unwrap()
    };

    Ok(re.replace_all(&result, &replacement).into_owned())
}

fn main() {
    // Create module structure from prost output
    let mut tree = Tree::default();
    fs::read_dir("protos").unwrap().for_each(|entry| {
        let path = entry.unwrap().path();
        let stem = path.file_stem().unwrap().to_str().unwrap().to_string();
        let parts: Vec<&str> = stem.split(".").collect();
        if parts.last() == Some(&"tonic") {
            return;
        }
        let filename = path.file_name().unwrap().to_str().unwrap().to_string();
        tree.push(&parts, filename);
    });

    let mut out_dir = PathBuf::from(env::var_os("OUT_DIR").unwrap());
    out_dir.push("proto.rs");
    fs::write(out_dir, tree.compile()).unwrap();
}
