use actix_web::{web, App, HttpRequest, HttpResponse, HttpServer};
use serde::{Deserialize, Serialize};
use std::sync::Arc;
use tokio::sync::Mutex;
use tonic::transport::Channel;

mod pb {
    tonic::include_proto!("weathertweet");
}

use pb::weather_tweet_service_client::WeatherTweetServiceClient;
use pb::{WeatherTweetRequest, Municipalities, Weathers};

#[derive(Serialize, Deserialize, Clone)]
struct WeatherTweet {
    municipality: String,
    temperature: i32,
    humidity: i32,
    weather: String,
}

struct AppState {
    grpc_client: Arc<Mutex<Option<WeatherTweetServiceClient<Channel>>>>,
}

#[actix_web::post("/api/tweets")]
async fn receive_tweet(
    req: HttpRequest,
    body: web::Json<WeatherTweet>,
    data: web::Data<AppState>,
) -> HttpResponse {
    log::info!("Recibido tweet: {:?}", body);

    let municipality_enum = match body.municipality.to_lowercase().as_str() {
        "mixco" => Municipalities::Mixco as i32,
        "guatemala" => Municipalities::Guatemala as i32,
        "amatitlan" => Municipalities::Amatitlan as i32,
        "chinautla" => Municipalities::Chinautla as i32,
        _ => Municipalities::MunicipalitiesUnknown as i32,
    };

    let weather_enum = match body.weather.to_lowercase().as_str() {
        "sunny" => Weathers::Sunny as i32,
        "cloudy" => Weathers::Cloudy as i32,
        "rainy" => Weathers::Rainy as i32,
        "foggy" => Weathers::Foggy as i32,
        _ => Weathers::WeathersUnknown as i32,
    };

    let request = WeatherTweetRequest {
        municipality: municipality_enum,
        temperature: body.temperature,
        humidity: body.humidity,
        weather: weather_enum,
    };

    match data.grpc_client.lock().await.as_mut() {
        Some(client) => {
            match client.send_tweet(request).await {
                Ok(response) => {
                    log::info!("Response: {}", response.get_ref().status);
                    HttpResponse::Ok().json(serde_json::json!({
                        "status": "success",
                        "message": response.get_ref().status.clone()
                    }))
                }
                Err(e) => {
                    log::error!("Error enviando a gRPC: {}", e);
                    HttpResponse::InternalServerError().json(serde_json::json!({
                        "status": "error",
                        "message": format!("Error: {}", e)
                    }))
                }
            }
        }
        None => {
            log::error!("Cliente gRPC no disponible");
            HttpResponse::ServiceUnavailable().json(serde_json::json!({
                "status": "error",
                "message": "gRPC service unavailable"
            }))
        }
    }
}

#[actix_web::get("/health")]
async fn health_check() -> HttpResponse {
    HttpResponse::Ok().json(serde_json::json!({"status": "ok"}))
}

#[actix_web::main]
async fn main() -> std::io::Result<()> {
    env_logger::init_from_env(env_logger::Env::new().default_filter_or("info"));

    let grpc_addr = std::env::var("GRPC_SERVER_ADDR")
        .unwrap_or_else(|_| "http://go-deployment-1:50051".to_string());

    log::info!("Conectando a gRPC en: {}", grpc_addr);

    let channel = Channel::from_shared(grpc_addr)
        .expect("URL inválida")
        .connect()
        .await
        .expect("No se pudo conectar al servidor gRPC");

    let client = WeatherTweetServiceClient::new(channel);

    let app_state = web::Data::new(AppState {
        grpc_client: Arc::new(Mutex::new(Some(client))),
    });

    log::info!("Iniciando servidor en 0.0.0.0:8080");

    HttpServer::new(move || {
        App::new()
            .app_data(app_state.clone())
            .service(receive_tweet)
            .service(health_check)
    })
    .bind("0.0.0.0:8080")?
    .run()
    .await
}
