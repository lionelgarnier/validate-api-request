# Validate API Request

A Go library to validate API requests against OpenAPI/Swagger specifications (v3).

## Overview

This library helps validate HTTP API requests by checking them against OpenAPI 3.0 specifications. It ensures incoming requests match the expected schemas, parameters, and formats defined in your API documentation.

## Features

- Validates requests against OpenAPI 3.0 specifications 
- Supports JSON and YAML specification formats
- Validates:
    - Request parameters
    - Request bodies
    - Response formats
    - Data types and formats
- Easy integration with Go web applications

## Installation

```bash
go get github.com/lionelgarnier/validate-api-request
```

## Usage

```go
import "github.com/lionelgarnier/validate-api-request"

// Load the OpenAPI spec
validator, err := validator.New("path/to/openapi.yaml")
if err != nil {
        log.Fatal(err)
}

// Validate a request
err = validator.ValidateRequest(request)
if err != nil {
        // Handle validation error
}
```

## Documentation

For detailed documentation and examples, please visit the [project wiki](https://github.com/lionelgarnier/validate-api-request/wiki).

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.