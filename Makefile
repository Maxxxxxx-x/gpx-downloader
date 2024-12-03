#!make

main_bin_name = gpx-downloader
main_cmd_path = ./cmd/${main_bin_name}

migration_bin_name = migrate
migration_cmd_path = ./cmd/${migration_bin_name}

tidy:
	go mod tidy
	go fmt ./...


clean:
	@if [ -d ./tmp ]; then rm -rf ./tmp; fi
	@if [ -d ./tmp/bin/${main_bin_name} ]; then rm -rf /tmp/bin/${main_bin_name}; fi


build/test: clean
	export APP_ENV=dev && go build -o=./tmp/bin/${main_bin_name} ${main_cmd_path}

build/migrate: clean
	@go build -o=/tmp/bin/${migration_bin_name} ${migration_cmd_path}

build/migrate-test: clean
	@go build -o=./tmp/bin/${migration_bin_name} ${migration_cmd_path}

build/prod: clean
	@go build -o=/tmp/bin/${main_bin_name} ${main_cmd_path}


test: build/test
	./tmp/bin/${main_bin_name}

migrate: build/migrate-test
	./tmp/bin/${migration_bin_name}
