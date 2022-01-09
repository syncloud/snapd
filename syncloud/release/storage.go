package main

type Storage interface {
	UploadFile(from string, to string) error
	UploadContent(content string, to string) error
}
