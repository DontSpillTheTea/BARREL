module github.com/DontSpillTheTea/barrel/apps/api

go 1.25.0

require gopkg.in/yaml.v3 v3.0.1

require (
	github.com/Azure/azure-sdk-for-go/sdk/data/aztables v1.4.1
	github.com/Azure/azure-sdk-for-go/sdk/storage/azblob v1.7.0
	github.com/google/uuid v1.6.0
)

require (
	github.com/Azure/azure-sdk-for-go/sdk/azcore v1.21.1 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/internal v1.12.0 // indirect
	golang.org/x/net v0.54.0 // indirect
	golang.org/x/text v0.37.0 // indirect
)
