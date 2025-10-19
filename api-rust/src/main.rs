use actix_web::{web, App, HttpServer, Responder, HttpResponse};
use serde::Deserialize;
use tonic::transport::Channel;

pub mod weather_tweet {
    tonic::include_proto!("wethertweet");
}

use weather_tweet::{weather_tweet_service_client::WeatherTweetServiceClient, WeatherTweetRequest};

#[derive(Deserialize, Debug)]
struct Tweet {
    municipality: String,
    temperature: i32,
    humidity: i32,
    weather: String,
}

fn map_municipality(municipality: &str) -> i32 {
    match municipality.to_lowercase().as_str() {
        "mixco" => 1,
        "guatemala" => 2,
        "amatitlan" => 3,
        "chinautla" => 4,
        _ => 0,
    }
}

fn map_weather(weather: &str) -> i32 {
    match weather.to_lowercase().as_str() {
        "sunny" => 1,
        "cloudy" => 2,
        "rainy" => 3,
        "foggy" => 4,
        _ => 0,
    }
}

async fn send_tweet(tweet: web::Json<Tweet>, data: web::Data<AppState>) -> impl Responder {
    println!("Received tweet: {:?}", tweet);

    let request = tonic::Request::new(WeatherTweetRequest {
        municipality: map_municipality(&tweet.municipality),
        temperature: tweet.temperature,
        humidity: tweet.humidity,
        weather: map_weather(&tweet.weather),
    });

    let mut client = data.grpc_client.clone();

    match client.send_tweet(request).await {
        Ok(response) => {
            println!("gRPC response: {:?}", response);
            HttpResponse::Ok().json(response.into_inner())
        }
        Err(e) => {
            eprintln!("gRPC request failed: {:?}", e);
            HttpResponse::InternalServerError().body("Failed to send tweet to Go service")
        }
    }
}

struct AppState {
    grpc_client: WeatherTweetServiceClient<Channel>,
}

#[actix_web::main]
async fn main() -> std::io::Result<()> {
    // TODO: Replace "go-service-grpc:50051" with the actual address of your Go gRPC service in Kubernetes
    let go_service_address = "http://go-service-grpc:50051";
    
    let grpc_client = WeatherTweetServiceClient::connect(go_service_address)
        .await
        .expect("Failed to connect to gRPC server");

    let app_state = web::Data::new(AppState {
        grpc_client,
    });

    println!("Server running at http://0.0.0.0:8080");

    HttpServer::new(move || {
        App::new()
            .app_data(app_state.clone())
            .route("/tweet", web::post().to(send_tweet))
    })
    .bind("0.0.0.0:8080")?
    .run()
    .await
}
