use actix_web::{web, App, HttpServer, HttpResponse, Responder};
use serde::{Deserialize, Serialize};
use reqwest;

#[derive(Debug, Deserialize, Serialize)]
struct TweetData {
    descripcion: String,
    country: String,
    weather: i32,
}

async fn handle_tweet(tweet: web::Json<TweetData>) -> impl Responder {
    println!("Received Tweet request: {:?}", tweet);

    // Conexion a go
    let go_service_url = "http://go-client-service:8080/grpc-go";

    // Enviar los datos a Go
    let client = reqwest::Client::new();

    match client.post(go_service_url)
        .json(&tweet.into_inner())
        .send()
        .await {
            Ok(response) => {
                if response.status().is_success() {
                    let response_body = response.text().await.unwrap_or_default();
                    HttpResponse::Ok().json(response_body)
                } else {
                    let error_msg = response.text().await.unwrap_or_else(|_| "Error desconocido".to_string());
                    HttpResponse::InternalServerError().json(error_msg)
                }
            },
            Err(e) => {
                HttpResponse::InternalServerError().json(format!("Error al enviar los datos a Go: {}", e))
            }
        }
}

#[actix_web::main]
async fn main() -> std::io::Result<()> {
    println!("Starting Rust HTTP server at http://0.0.0.0:8081");
    
    HttpServer::new(|| {
        App::new()
            .route("/input", web::post().to(handle_tweet))
    })
    .bind("0.0.0.0:8081")?
    .run()
    .await
}