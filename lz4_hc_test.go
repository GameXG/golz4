// Package lz4 implements compression using lz4.c. This is its test
// suite.
//
// Copyright (c) 2013 CloudFlare, Inc.

package lz4

import (
	"io/ioutil"
	"reflect"
	"strings"
	"testing"
	"testing/quick"
)

func TestCompressionHCRatio(t *testing.T) {
	input, err := ioutil.ReadFile("sample.txt")
	if err != nil {
		t.Fatal(err)
	}
	output := make([]byte, CompressBound(input))
	outSize, err := CompressHC(input, output)
	if err != nil {
		t.Fatal(err)
	}

	if want := 4354; want != outSize {
		t.Fatalf("HC Compressed output length != expected: %d != %d", want, outSize)
	}
}

func TestCompressionHCLevels(t *testing.T) {
	input, err := ioutil.ReadFile("sample.txt")
	if err != nil {
		t.Fatal(err)
	}

	data := make([]byte, len(input))

	cases := []struct {
		Level   int
		Outsize int
	}{
		{0, 4354},
		{1, 4450},
		{2, 4395},
		{3, 4372},
		{4, 4358},
		{5, 4354},
		{6, 4354},
		{7, 4354},
		{8, 4354},
		{9, 4354},
		{10, 4354},
		{11, 4351},
		{12, 4351},
		{13, 4351},
		{14, 4351},
		{15, 4351},
		{16, 4351},
	}

	for _, tt := range cases {
		output := make([]byte, CompressBound(input))
		outSize, err := CompressHCLevel(input, output, tt.Level)
		if err != nil {
			t.Fatal(err)
		}

		if want := tt.Outsize; want != outSize {
			t.Errorf("HC level %d length != expected: %d != %d",
				tt.Level, want, outSize)
		}

		err = Uncompress(output[:outSize], data)
		if err != nil {
			t.Error("[Uncompress]HC level %d ,%v", tt.Level, err)
		}

		if !reflect.DeepEqual(input, data) {
			t.Error("input!=data")
		}
	}
}

func TestCompressionHC(t *testing.T) {
	input := []byte(strings.Repeat("Hello world, this is quite something", 10))
	output := make([]byte, CompressBound(input))
	outSize, err := CompressHC(input, output)
	if err != nil {
		t.Fatalf("Compression failed: %v", err)
	}
	if outSize == 0 {
		t.Fatal("Output buffer is empty.")
	}
	output = output[:outSize]
	decompressed := make([]byte, len(input))
	err = Uncompress(output, decompressed)
	if err != nil {
		t.Fatalf("Decompression failed: %v", err)
	}
	if string(decompressed) != string(input) {
		t.Fatalf("Decompressed output != input: %q != %q", decompressed, input)
	}
}

func TestEmptyCompressionHC(t *testing.T) {
	input := []byte("")
	output := make([]byte, CompressBound(input))

	outSize, err := CompressHC(input, output)
	if err != nil {
		t.Fatalf("Compression failed: %v", err)
	}
	if outSize == 0 {
		t.Fatal("Output buffer is empty.")
	}
	output = output[:outSize]
	decompressed := make([]byte, len(input))
	err = Uncompress(output, decompressed)
	if err != nil {
		t.Fatalf("Decompression failed: %v", err)
	}
	if string(decompressed) != string(input) {
		t.Fatalf("Decompressed output != input: %q != %q", decompressed, input)
	}
}

func TestNoCompressionHC(t *testing.T) {
	input := []byte("ABCDEFGHIJKLMNOPQRSTUVWXYZ")
	output := make([]byte, CompressBound(input))
	outSize, err := CompressHC(input, output)
	if err != nil {
		t.Fatalf("Compression failed: %v", err)
	}
	if outSize == 0 {
		t.Fatal("Output buffer is empty.")
	}
	output = output[:outSize]
	decompressed := make([]byte, len(input))
	err = Uncompress(output, decompressed)
	if err != nil {
		t.Fatalf("Decompression failed: %v", err)
	}
	if string(decompressed) != string(input) {
		t.Fatalf("Decompressed output != input: %q != %q", decompressed, input)
	}
}

func TestCompressionErrorHC(t *testing.T) {
	input := []byte(strings.Repeat("Hello world, this is quite something", 10))
	output := make([]byte, 0)
	outSize, err := CompressHC(input, output)

	if outSize != 0 {
		t.Fatalf("%d", outSize)
	}

	if err == nil {
		t.Fatalf("Compression should have failed but didn't")
	}

	output = make([]byte, 1)
	_, err = CompressHC(input, output)
	if err == nil {
		t.Fatalf("Compression should have failed but didn't")
	}
}

func TestDecompressionErrorHC(t *testing.T) {
	input := []byte(strings.Repeat("Hello world, this is quite something", 10))
	output := make([]byte, CompressBound(input))
	outSize, err := CompressHC(input, output)
	if err != nil {
		t.Fatalf("Compression failed: %v", err)
	}
	if outSize == 0 {
		t.Fatal("Output buffer is empty.")
	}
	output = output[:outSize]
	decompressed := make([]byte, len(input)-1)
	err = Uncompress(output, decompressed)
	if err == nil {
		t.Fatalf("Decompression should have failed")
	}

	decompressed = make([]byte, 1)
	err = Uncompress(output, decompressed)
	if err == nil {
		t.Fatalf("Decompression should have failed")
	}

	decompressed = make([]byte, 0)
	err = Uncompress(output, decompressed)
	if err == nil {
		t.Fatalf("Decompression should have failed")
	}
}

func TestFuzzHC(t *testing.T) {
	f := func(input []byte) bool {
		output := make([]byte, CompressBound(input))
		outSize, err := CompressHC(input, output)
		if err != nil {
			t.Fatalf("Compression failed: %v", err)
		}
		if outSize == 0 {
			t.Fatal("Output buffer is empty.")
		}
		output = output[:outSize]
		decompressed := make([]byte, len(input))
		err = Uncompress(output, decompressed)
		if err != nil {
			t.Fatalf("Decompression failed: %v", err)
		}
		if string(decompressed) != string(input) {
			t.Fatalf("Decompressed output != input: %q != %q", decompressed, input)
		}

		return true
	}

	conf := &quick.Config{MaxCount: 20000}
	if testing.Short() {
		conf.MaxCount = 1000
	}
	if err := quick.Check(f, conf); err != nil {
		t.Fatal(err)
	}
}
