#!/usr/bin/env bash


path=.
main_file=../main.go
	
platforms=("windows/amd64" "windows/386" "linux/amd64" "linux/386" "linux/arm" "linux/arm64" "wasip1/wasm" "js/wasm")

for platform in "${platforms[@]}"
do
	platform_split=(${platform//\// })
	GOOS=${platform_split[0]}
	GOARCH=${platform_split[1]}
	output_name=$path'/'$GOOS'-'$GOARCH
	if [ $GOOS = "windows" ]; then
		output_name+='.exe'
	elif [ $GOARCH = "wasm" ]; then
		output_name=$path'/'$GOOS'.wasm'
	else
		output_name+='.bin'
	fi	
	env GOOS=$GOOS GOARCH=$GOARCH go build -o $output_name $main_file
	if [ $? -ne 0 ]; then
   		echo 'An error has occurred! '
		echo 'GOOS='$GOOS
		echo 'GOARCH='$GOARCH
	fi
done