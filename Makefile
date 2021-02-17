export TOP_DIR:=$(dir $(realpath $(firstword $(MAKEFILE_LIST))))

stressor-go:
	mkdir -p $(TOP_DIR)/out; \
	go build -ldflags='-extldflags=-static' -o $(TOP_DIR)/out/stressor-go golang/stressor/stressor.go 

stressor-rust:
	export RUSTFLAGS="-C target-feature=+crt-static"; \
	mkdir -p $(TOP_DIR)/out; \
	cd rust/stressor; \
	cargo build --release; \
	mv target/release/stressor-rust $(TOP_DIR)/out/

.PHONY: all
all: stressor-go stressor-rust

.DEFAULT_GOAL := all
