use actix_web::{web, App, HttpResponse, HttpServer, Responder};
use log::{error, info};
use reqwest::Client;
use serde::{Deserialize, Serialize};
use std::env;

// Definición de enumeraciones para municipios y clima
#[derive(Debug, Serialize, Deserialize, Clone, Copy)]
enum Municipality {
    #[serde(rename = "mixco")]
    Mixco,
    #[serde(rename = "guatemala")]
    Guatemala,
    #[serde(rename = "amatitlan")]
    Amatitlan,
    #[serde(rename = "chinautla")]
    Chinautla,
}

#[derive(Debug, Serialize, Deserialize, Clone, Copy)]
enum Weather {
    #[serde(rename = "sunny")]
    Sunny,
    #[serde(rename = "cloudy")]
    Cloudy,
    #[serde(rename = "rainy")]
    Rainy,
    #[serde(rename = "foggy")]
    Foggy,
}

// Estructura del tweet del clima
#[derive(Debug, Serialize, Deserialize)]
struct WeatherTweet {
    municipality: Municipality,
    temperature: i32,
    humidity: i32,
    weather: Weather,
}

// Respuesta para el cliente
#[derive(Debug, Serialize)]
struct ApiResponse {
    status: String,
    message: String,
}

// Controlador para el endpoint de tweets
async fn tweet(tweet_data: web::Json<WeatherTweet>, client: web::Data<Client>) -> impl Responder {
    let go_service_url = env::var("GO_SERVICE_URL").unwrap_or_else(|_| "http://go-service:8080/api/v1/tweet".to_string());

    info!("Recibido tweet: {:?}", tweet_data);

    // Reenviar datos al servicio Go
    match client.post(&go_service_url)
        .json(&tweet_data)
        .send()
        .await {
            Ok(response) => {
                if response.status().is_success() {
                    info!("Tweet enviado correctamente al servicio Go");
                    HttpResponse::Ok().json(ApiResponse {
                        status: "success".to_string(),
                        message: "Tweet procesado correctamente".to_string(),
                    })
                } else {
                    let status = response.status();
                    error!("Error al enviar tweet al servicio Go. Código: {}", status);
                    HttpResponse::InternalServerError().json(ApiResponse {
                        status: "error".to_string(),
                        message: format!("Error al procesar tweet: {}", status),
                    })
                }
            },
            Err(err) => {
                error!("Error de conexión con el servicio Go: {}", err);
                HttpResponse::InternalServerError().json(ApiResponse {
                    status: "error".to_string(),
                    message: format!("Error de conexión con el servicio interno: {}", err),
                })
            }
        }
}

// Endpoint de estado para health checks
async fn health() -> impl Responder {
    HttpResponse::Ok().json(ApiResponse {
        status: "success".to_string(),
        message: "API funcionando correctamente - 202200129".to_string(),
    })
}

#[actix_web::main]
async fn main() -> std::io::Result<()> {
    // Configuración de logs
    env_logger::init_from_env(env_logger::Env::default().default_filter_or("info"));

    // Puerto de la API
    let port = env::var("PORT").unwrap_or_else(|_| "8000".to_string());
    let addr = format!("0.0.0.0:{}", port);

    info!("Iniciando servidor en {}", addr);

    // Cliente HTTP para comunicarse con el servicio Go
    let client = Client::new();

    // Configuración del servidor HTTP
    HttpServer::new(move || {
        App::new()
            .app_data(web::Data::new(client.clone()))
            .service(
                web::scope("/api/v1")
                    .route("/tweet", web::post().to(tweet))
            )
            .route("/health", web::get().to(health))
    })
    .bind(addr)?
    .workers(4) // Número de workers para manejar concurrencia
    .run()
    .await
}