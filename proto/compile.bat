protoc -I. -I"C:\Development\googleapis" --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=require_unimplemented_servers=false,paths=source_relative recipesvc.proto
protoc -I. -I"C:\Development\googleapis" --grpc-gateway_out=. --grpc-gateway_opt=logtostderr=true --grpc-gateway_opt=paths=source_relative recipesvc.proto
protoc -I. -I"C:\Development\googleapis" --openapiv2_out=logtostderr=true:. --openapiv2_opt=output_format=yaml recipesvc.proto