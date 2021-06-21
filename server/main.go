package main

import (
	"context"
	"fmt"
	"github.com/Cloudwalker-Technologies/JwtsearchService/db_manager"
	pb "github.com/Cloudwalker-Technologies/JwtsearchService/proto"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/joho/godotenv"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
)

var (
	mongoDbUrl           string
	grpcPort             string
	restPort             string
	authHost             string
	authPort             string
	contentDbName 		 string
	contentCollectionName string
)

// loading .env keys and its values
func loadEnv() {
	mongoDbUrl = os.Getenv("MONGO_DB_URL")
	log.Println(mongoDbUrl)
	grpcPort = os.Getenv("GRPC_PORT")
	log.Println(grpcPort)
	restPort = os.Getenv("REST_PORT")
	log.Println(restPort)
	authHost = os.Getenv("AUTH_HOST")
	log.Printf(authHost)
	authPort = os.Getenv("AUTH_PORT")
	log.Printf(authPort)
	contentDbName = os.Getenv("CONTENT_DB_NAME")
	log.Printf(contentDbName)
	contentCollectionName = os.Getenv("CONTENT_COLLECTION_NAME")
	log.Printf(contentCollectionName)

}

// Init function to load .env file and grab values in it.
func initializeProcess() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	loadEnv()
}

func init() {
	println("init 1.0.0")
	initializeProcess()
}

func main() {
	go startGRPCServer(grpcPort)
	go startRESTServer(restPort, grpcPort)
	select {}
}

// starting a grpc server
func startGRPCServer(address string) error {
	lis, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}

	// Create an array of gRPC options with the credentials
	//opts := []grpc.ServerOption{grpc.UnaryInterceptor(unaryInterceptor)}
	opts := []grpc.ServerOption{}

	// create a gRPC server object
	s := grpc.NewServer(opts...)
	contentCollection := db_manager.GetMongoDbCollection(mongoDbUrl, contentDbName, contentCollectionName)

	pb.RegisterJwtsearchServiceServer(s, &jwserver{
		contentCollection,
	})
	return s.Serve(lis)
}

// starting a rest server using grpc-rest.
func startRESTServer(address, grpcAddress string) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	mux := runtime.NewServeMux(
		runtime.WithIncomingHeaderMatcher(CustomMatcher),
		runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{
			MarshalOptions: protojson.MarshalOptions{
				UseProtoNames:   true,
				UseEnumNumbers:  true,
				EmitUnpopulated: false,
			},
			UnmarshalOptions: protojson.UnmarshalOptions{},
		}),
		runtime.WithForwardResponseOption(httpResponseModifier),
	)
	opts := []grpc.DialOption{grpc.WithInsecure()} // Register ping
	err := pb.RegisterJwtsearchServiceHandlerFromEndpoint(ctx, mux, grpcAddress, opts)
	if err != nil {
		return fmt.Errorf("could not register service Ping: %s", err)
	}
	log.Printf("starting HTTP/1.1 REST server on %s", address)
	log.Printf("starting HTTP/2 GRPC server on %s", grpcAddress)
	return http.ListenAndServe(address, mux)
}

/*
func unaryInterceptor(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler) (interface{}, error) {

	s, ok := info.Server.(*uiconfigServer)
	if !ok {
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("unable to cast the server"))
	}
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "metadata is not provided")
	}
	values := md["authorization"]
	if len(values) == 0 {
		return nil, status.Errorf(codes.Unauthenticated, "authorization token is not provided")
	}
	accessToken := values[0]
	outGoingContext := metadata.NewOutgoingContext(context.Background(), md)
	authResp, err := pbAuth.NewAuthServiceClient(s.authServiceConn).ValidateToken(outGoingContext, &pbAuth.TokenRequest{Token: accessToken})
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "access token is invalid: %v", err)
	}
	if authResp.AuthState == pbAuth.AUTHORIZATION_STATE_NO_AUTHORIZED {
		return nil, status.Errorf(codes.Unauthenticated, "access token is invalid: %v", err)
	}
	return handler(ctx, req)
}

*/

func makingAuthServiceConnection() *grpc.ClientConn {
	grpcAuthUrl := fmt.Sprintf("%s%s", authHost, authPort)
	log.Println(grpcAuthUrl)
	conn, err := grpc.Dial(grpcAuthUrl, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %s", err)
	}
	return conn
}

// CustomMatcher allow all the REST headers for grpc-gateway
func CustomMatcher(key string) (string, bool) {
	return key, true
}

func httpResponseModifier(ctx context.Context, w http.ResponseWriter, p proto.Message) error {
	md, ok := runtime.ServerMetadataFromContext(ctx)
	if !ok {
		return nil
	}
	// set http status code
	if vals := md.HeaderMD.Get("x-http-code"); len(vals) > 0 {
		code, err := strconv.Atoi(vals[0])
		if err != nil {
			return err
		}
		w.WriteHeader(code)
	}
	return nil
}
