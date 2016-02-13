all: install

checkdeps:
	@echo "Checking deps:"
	@(env bash $(PWD)/buildscripts/checkdeps.sh)

checkgopath:
	@echo "Checking if project is at ${GOPATH}"
	@for minioxlpath in $(echo ${GOPATH} | sed 's/:/\n/g'); do if [ ! -d ${minioxlpath}/src/github.com/minio/minio-xl ]; then echo "Project not found in ${minioxlpath}, please follow instructions provided at https://github.com/minio/minio-xl/blob/master/CONTRIBUTING.md#setup-your-minio-github-repository" && exit 1; fi done

getdeps: checkdeps checkgopath
	@go get -u github.com/golang/lint/golint && echo "Installed golint:"
	@go get -u golang.org/x/tools/cmd/vet && echo "Installed vet:"
	@go get -u github.com/fzipp/gocyclo && echo "Installed gocyclo:"
	@go get -u github.com/remyoudompheng/go-misc/deadcode && echo "Installed deadcode:"

verifiers: getdeps vet fmt lint cyclo

vet:
	@echo "Running $@:"
	@GO15VENDOREXPERIMENT=1 go tool vet -all *.go
	@GO15VENDOREXPERIMENT=1 go tool vet -all ./pkg
	@GO15VENDOREXPERIMENT=1 go tool vet -shadow=true *.go
	@GO15VENDOREXPERIMENT=1 go tool vet -shadow=true ./pkg

fmt:
	@echo "Running $@:"
	@GO15VENDOREXPERIMENT=1 gofmt -s -l *.go
	@GO15VENDOREXPERIMENT=1 gofmt -s -l pkg

lint:
	@echo "Running $@:"
	@GO15VENDOREXPERIMENT=1 golint *.go
	@GO15VENDOREXPERIMENT=1 golint github.com/minio/minio-xl/pkg...

cyclo:
	@echo "Running $@:"
	@GO15VENDOREXPERIMENT=1 gocyclo -over 25 *.go
	@GO15VENDOREXPERIMENT=1 gocyclo -over 25 pkg

build: constants getdeps verifiers
	@GO15VENDOREXPERIMENT=1 go generate ./...

isa-l:
	@echo "Configuring $@:"
	@git clone -q https://github.com/minio/isa-l.git
	@mkdir -p build
	@cd build; $(PWD)/isa-l/configure --prefix $(PWD)/build/lib --sysconfdir $(PWD)/build/lib --includedir $(PWD)/build/lib --libdir $(PWD)/build/lib >/dev/null
	@make -C build >/dev/null
	@make -C build install >/dev/null

test: build isa-l
	@echo "Running all testing:"
	@CGO_CPPFLAGS="-I$(PWD)/build/lib" CGO_LDFLAGS="$(PWD)/build/lib/libisal.a" GO15VENDOREXPERIMENT=1 go test $(GOFLAGS) .
	@CGO_CPPFLAGS="-I$(PWD)/build/lib" CGO_LDFLAGS="$(PWD)/build/lib/libisal.a" GO15VENDOREXPERIMENT=1 go test $(GOFLAGS) github.com/minio/minio-xl/pkg...

gomake-all: build isa-l
	@echo "Installing minio-xl:"
	@CGO_CPPFLAGS="-I$(PWD)/build/lib" CGO_LDFLAGS="$(PWD)/build/lib/libisal.a" GO15VENDOREXPERIMENT=1 go install github.com/minio/minio-xl

constants:
	@echo "Generating new build-constants.go"
	@GO15VENDOREXPERIMENT=1 go run buildscripts/gen-constants.go

pkg-add:
	@GO15VENDOREXPERIMENT=1 govendor add $(PKG)

pkg-update:
	@GO15VENDOREXPERIMENT=1 govendor update $(PKG)

pkg-remove:
	@GO15VENDOREXPERIMENT=1 govendor remove $(PKG)

install: gomake-all

clean:
	@echo "Cleaning up all the generated files:"
	@rm -fv build-constants.go
	@rm -fv cover.out
	@rm -fv minio-xl
	@rm -rf isa-l
	@rm -rf build
