use openssl::ssl::{select_next_proto, AlpnError, SslAcceptor, SslFiletype, SslMethod};
use proto::{
    greeter_server::{Greeter, GreeterServer},
    HelloReply, HelloRequest,
};
use std::env;
use std::net::SocketAddr;
use tokio::net::TcpListener;
use tokio_stream::wrappers::TcpListenerStream;
use tonic::{
    transport::{server::TcpConnectInfo, Server},
    Request, Response, Status,
};
use tonic_openssl::{SslConnectInfo, ALPN_H2_WIRE};

mod proto {
    tonic::include_proto!("helloworld");
}

#[derive(Debug, Default)]
pub struct GreeterService {}

#[tonic::async_trait]
impl Greeter for GreeterService {
    async fn say_hello(
        &self,
        request: Request<HelloRequest>,
    ) -> Result<Response<HelloReply>, Status> {
        let remote_addr = request
            .extensions()
            .get::<SslConnectInfo<TcpConnectInfo>>()
            .and_then(|info| info.get_ref().remote_addr());
        println!("Got a request from {:?}", remote_addr);

        println!("Request from client: {:?} with info", request.remote_addr());

        let reply = HelloReply {
            message: format!("Server says hello {}!", request.into_inner().name),
        };

        Ok(Response::new(reply))
    }
}

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    let current_dir = env::current_dir()?;

    let file = match current_dir.ends_with("signal") {
        true => "tls/server",
        false => "signal/tls/server",
    };

    let private_key = current_dir.join(file.to_owned() + ".key");
    let certificate = current_dir.join(file.to_owned() + ".crt");

    // Setup TLS
    let mut acceptor = SslAcceptor::mozilla_intermediate(SslMethod::tls()).unwrap();
    acceptor
        .set_private_key_file(private_key, SslFiletype::PEM)
        .unwrap();
    acceptor.set_certificate_chain_file(certificate).unwrap();

    acceptor.check_private_key().unwrap();
    acceptor.set_alpn_protos(ALPN_H2_WIRE)?;
    acceptor.set_alpn_select_callback(|_ssl, alpn| {
        select_next_proto(ALPN_H2_WIRE, alpn).ok_or(AlpnError::NOACK)
    });

    let acceptor = acceptor.build();

    // Setup server
    let addr = "0.0.0.0:8080".parse::<SocketAddr>()?;

    let listener = TcpListener::bind(addr).await?;
    let incoming = tonic_openssl::incoming(TcpListenerStream::new(listener), acceptor);

    let greeter = GreeterService::default();

    Server::builder()
        .add_service(GreeterServer::new(greeter))
        .serve_with_incoming(incoming)
        .await?;
    Ok(())
}
