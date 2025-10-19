fn main() {
    tonic_build::compile_protos("../proto/weather_tweet.proto")
        .expect("Failed to compile protos");
}
