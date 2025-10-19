fn main() {
    tonic_build::compile_protos("../proto/weather_tweet.proto").unwrap();
}
