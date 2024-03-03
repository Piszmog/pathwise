# Rust + HTMX Template

This is a template repository for stiches together a number of technologies that you may use 
to server a website using Rust and HTMX.

## Getting Started

In the top right, select the dropdown __Use this template__ and select __Create a new repository__.

## Technologies

A few different technologies are configured to help getting off the ground easier.

- [tokio-rusqlite](https://github.com/programatik29/tokio-rusqlite/tree/master) for database layer
  - Stubbed to use SQLite
- [Tailwind CSS](https://tailwindcss.com/) for styling
  - Output is generated with the [CLI](https://tailwindcss.com/docs/installation)
  - Automatically ran when the application is built (set in `build.rs`)
- [Askama](https://djc.github.io/askama/askama.html) for creating HTML
- [HTMX](https://htmx.org/) for HTML interaction
- [cargo-watch](https://github.com/watchexec/cargo-watch) to rebuild your application when something changes.

## Structure

```text
.
├── Cargo.lock
├── Cargo.toml
├── README.md
├── assets
│   ├── css
│   │   └── output@dev.css
│   ├── img
│   │   └── favicon.ico
│   ├── js
│   │   └── htmx@1.9.10.min.js
│   └── styles
│       └── input.css
├── build.rs
├── db.sqlite
├── src
│   ├── db
│   │   └── mod.rs
│   └── main.rs
├── tailwind.config.js
└── templates
    ├── base.html
    ├── heading.html
    └── home.html
```

### Templates

This is where you create the Askama templates to be rendered.

### DB

General module where to put your database releated operations.

### Assets

This is where your assets live. Any Javascript, images, or styling needs to go in the 
`assets` directory. The directory will be embedded into the application.

Note, the `assets/css` will be ignored by `git` (configured in `.gitignore`) since the 
files that are written to this directory are done by the Tailwind CSS CLI. Custom styles should
go in the `assets/styles/input.css` file.

### Styles

This contains the `input.css` that the Tailwind CSS CLI uses to generate your output CSS. 
Update `input.css` with any custom CSS you need and it will be included in the output CSS.

## Run

Running the application is straight forward.

### Prerequisites

- Install [tailwindcss CLI](https://tailwindcss.com/docs/installation)

### Cargo

Simply run the cargo command to run the application.

```shell
cargo run
```

Or run `cargo-watch` to have the application rebuilt when something changes.

```shell
cargo watch -x run
```

## Github Workflow

The repository comes with two Github workflows as well. One called `ci.yml` that lints and 
tests your code. The other called `release.yml` that creates a tag, GitHub Release, and 
attaches the binaries to the Release.

