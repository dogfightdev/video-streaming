# builds the govod executable
cd ./govod

# remove previous build
rm ./build/govod

go build -o ./build/govod ./cmd/govod