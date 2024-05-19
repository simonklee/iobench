#!/usr/bin/env bash

set -o errexit
set -o pipefail
set -o nounset
#set -x

SCRIPT_DIR=$(dirname "$0")

run_fio() {
	local root_path="$1"
	local name="$2"
	local rw="$3"
	local bs="$4"
	local size="$5"
	local numjobs="$6"
	local runtime="$7"

	# Create a subdirectory for the test files
	local test_dir="$root_path/output/$name"
	mkdir -p "$test_dir"

	# Run fio benchmark and save output directly in the test directory
	fio --name="$name" \
		--directory="$test_dir" \
		--ioengine=libaio \
		--rw="$rw" \
		--bs="$bs" \
		--size="$size" \
		--numjobs="$numjobs" \
		--runtime="$runtime" \
		--group_reporting \
		--output="$test_dir/$name.json" \
		--output-format=json
}

parse_args() {
	# Set root_path to the first argument or default to current directory.
	root_path="${1:-./}"

	# Strip trailing slash.
	root_path="${root_path%/}"
}

check_dependencies() {
	for dep in fio go; do
		if ! command -v "$dep" &>/dev/null; then
			echo "$dep is not installed. Please install $dep before running benchmarks."
			exit 1
		fi
	done
}

run_benchmarks() {
	# Create output directory
	mkdir -p "$root_path/output"

	# Run benchmarks
	run_fio "$root_path" "random_read" "randread" "4k" "1G" "4" "60"
	run_fio "$root_path" "random_write" "randwrite" "4k" "1G" "4" "60"
	run_fio "$root_path" "seq_read" "read" "128k" "1G" "4" "60"
	run_fio "$root_path" "seq_write" "write" "128k" "1G" "4" "60"

	# Generate bandwidth plot
	go run "$SCRIPT_DIR/plot-results.go" -input="$root_path/output" -output="$root_path/output/bandwidth.png" -title="SSD Benchmark Bandwidth" -xlabel="Test Type" -ylabel="Bandwidth (MB/s)" -value-type="bw"

	# Generate IOPS plot
	go run "$SCRIPT_DIR/plot-results.go" -input="$root_path/output" -output="$root_path/output/iops.png" -title="SSD Benchmark IOPS" -xlabel="Test Type" -ylabel="IOPS" -value-type="iops"
}

if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
	parse_args "$@"
	check_dependencies
	run_benchmarks
fi
