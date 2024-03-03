use askama::Template;
use askama_axum::{IntoResponse, Response};
use axum::{body, extract::State, http::StatusCode, response::Redirect, routing::get, Form};
use rust_embed::RustEmbed;
use serde::Deserialize;
use tokio_rusqlite::Connection;
use tower_http::compression::CompressionLayer;

mod db;

#[cfg(not(debug_assertions))]
const VERSION: &str = env!("CARGO_PKG_VERSION");

#[cfg(debug_assertions)]
const VERSION: &str = "dev";

#[derive(RustEmbed, Clone)]
#[folder = "assets/"]
struct Assets;

mod embedded {
    use refinery::embed_migrations;
    embed_migrations!("migrations");
}

#[tokio::main]
async fn main() {
    let mut migrations_conn = rusqlite::Connection::open("./db.sqlite").unwrap();
    embedded::migrations::runner()
        .run(&mut migrations_conn)
        .unwrap();

    let comression_layer = CompressionLayer::new().gzip(true);

    let app = axum::Router::new()
        .route(
            "/favicon.ico",
            get(|| async { Redirect::to("/assets/img/favicon.ico") }),
        )
        .route("/signup", get(signup))
        .route("/signin", get(signin))
        .nest_service("/assets", axum_embed::ServeEmbed::<Assets>::new())
        .layer(comression_layer)
        .with_state(Connection::open("./db.sqlite").await.unwrap());

    let listener = tokio::net::TcpListener::bind("0.0.0.0:8080").await.unwrap();
    axum::serve(listener, app).await.unwrap();
}

async fn signup() -> SignUpTemplate {
    SignUpTemplate {}
}

async fn register(Form(input): Form<SignUpForm>, State(state): State<Connection>) -> Response {
    if input.password != input.password_confirmation {
        // TODO
    }
    // if valid_pass
    // hash password
    // insert user
    ([("HX-Redirect", "/signin")]).into_response()
}

#[derive(Deserialize, Debug)]
struct SignUpForm {
    email: String,
    password: String,
    password_confirmation: String,
}

async fn signin() -> SignInTemplate {
    SignInTemplate {}
}

#[derive(Template)]
#[template(path = "signup.html")]
struct SignUpTemplate;

#[derive(Template)]
#[template(path = "signin.html")]
struct SignInTemplate;

#[derive(Template)]
#[template(path = "alert.html")]
struct AlertTemplate<'a> {
    alert_type: AlertType,
    title: &'a str,
    messages: &'a [&'a str],
}
#[derive(PartialEq)]
enum AlertType {
    Success,
    Warning,
    Error,
}
