# Wizard101 Multi-Client Manager

This project allows you to manage multiple instances of the Wizard101 game client. It automates the process of opening the clients, logging in with different accounts, and positioning the windows on your screen.

## Features

- Open multiple instances of Wizard101
- Automatically log in with different accounts
- Position each game window at specified coordinates

## Prerequisites

- Rust (https://www.rust-lang.org/tools/install)
- Wizard101 installed on your system

## Configuration

Create a `config.json` file in the root directory of the project with the following structure:

```json
{
    "file_path": "C:\\Path\\To\\Wizard101",
    "accounts_data": [
        {
            "username": "your_username_1",
            "password": "your_password_1",
            "x_pos": 100,
            "y_pos": 100
        },
        {
            "username": "your_username_2",
            "password": "your_password_2",
            "x_pos": 500,
            "y_pos": 100
        }
        // Add more accounts as needed
    ]
}
```