pub mod proto {
    include!(concat!(env!("OUT_DIR"), "/proto.rs"));
}

pub use prost::bytes;
