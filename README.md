# ShortURL Service
ShortURL Service is a web application that allows users to shorten long URLs into more manageable and shareable links.

[![codecov](https://codecov.io/gh/gennadis/shorturl/graph/badge.svg?token=4EUV9MONJG)](https://codecov.io/gh/gennadis/shorturl)

## Features
- **Shortening URLs**: Users can shorten long URLs into shorter, more manageable links.
- **User Management**: Authentication and authorization functionalities for managing users and their URLs.
- **Background Deletion**: Automatic background deletion of expired or unused URLs.
- #TODO: add more features

## Architecture
The ShortURL Service is built using the following technologies and architectural components:
- **Backend Language**: Go (Golang)
- **Database**: PostgreSQL
- **Storage Options**: Supports in-memory, file-based, and PostgreSQL storage for URLs data.
- **Web Framework**: Go-Chi for handling HTTP requests and routing.
- **Background Processing**: Uses goroutines for background deletion of URLs.

## Setup
To run the ShortURL Service locally, follow these steps:
1. Clone this repository to your local machine.
2. Install Go and PostgreSQL if you haven't already.
3. ...  
#TODO: add more setup details

## Usage
Once the ShortURL Service is running, you can use it as follows:
1. Use the provided APIs to programmatically interact with the service.
2. ...  
#TODO: add more usage details

## API Documentation
#TODO: implement Swagger support

## Contributing
Contributions are welcome! If you find any issues or have suggestions for improvements, please open an issue or submit a pull request.

## License
This project is licensed under the [MIT License](LICENSE).
