use rand::{Rng, distributions::Standard};
use std::thread;
use thread_id;
//use rand::distributions::Alphanumeric;

const MAXBUFFSIZE: i32 = 128 * 1024 * 1024;
const BUFFSIZE: i32 = 16;

fn populate_vectors() {
    let mut mvec: Vec<u8> = Vec::new();
    let mut done: i32 = 0;
    loop {
        if MAXBUFFSIZE - done < 16 { break; }

        mvec.append(&mut rand::thread_rng().sample_iter(Standard).take(BUFFSIZE as usize).collect());
        done = done + BUFFSIZE;
    }
    println!("Length of Vector {:?}", mvec.len());
    println!("{:?}", thread_id::get());
}

fn main() {

    // Use this to test a single core if you run into issues.
    //let mycore_id = core_affinity::CoreId{id: 1};
    //let handle = thread::spawn(move || {
    //    core_affinity::set_for_current(mycore_id);
    //    populate_vectors();
    //});
    //handle.join().unwrap();
    
    let core_ids = core_affinity::get_core_ids().unwrap();
    let handles = core_ids.into_iter().map(|id| {
        thread::spawn(move || {
            core_affinity::set_for_current(id);
            populate_vectors();
        })
    }).collect::<Vec<_>>();

    for handle in handles.into_iter() {
        handle.join().unwrap();
    }
    println!("Master thread :: {:?}", thread_id::get());
}

