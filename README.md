# Example Go Application with Solid Embed

This is an example Go application demonstrating the usage of embedding assets (such as templates, static files, etc.) into the binary using Go's embed package.

## Features

- Embedding of assets into the binary.
- Minimalistic setup with commands to build and run the application.

## Prerequisites

- Go installed on your system. You can download it from [here](https://golang.org/dl/).

## Setup

1. Clone this repository:

   ```bash
   git clone $THIS_REPO
   ```

2. Navigate to the project directory:

   ```bash
   cd $REPO
   ```

3. Build the application:

   ```bash
   make build-app
   ```

## Seed

- Using the this command will remove the old database and start a new one, already seeded:

  ```bash
  make seed
  ```

## Usage

- Run the application:

  ```bash
  ./bin/app
  ```

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
