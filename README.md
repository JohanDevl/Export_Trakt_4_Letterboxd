![Export_Trakt_4_Letterboxd](https://socialify.git.ci/u2pitchjami/Export_Trakt_4_Letterboxd/image?description=1&descriptionEditable=The%20purpose%20of%20this%20script%20is%20to%20export%20Trakt%20movies%20watchlist%20to%20csv%20file%20for%20manual%20Letterboxd%20import&font=Jost&language=1&logo=https%3A%2F%2Fgreen-berenice-35.tiiny.site%2Fimage2vector-3.svg&name=1&owner=1&pattern=Charlie%20Brown&stargazers=1&theme=Dark)

# Export Trakt 4 Letterboxd

This project allows you to export your Trakt.tv data to a format compatible with Letterboxd.

## Prerequisites

- A Trakt.tv account
- A Trakt.tv application (see below)
- jq (for JSON processing)
- curl (for API requests)

## Configuration

### 1. Create a Trakt.tv application

1. Log in to your Trakt.tv account
2. Go to https://trakt.tv/oauth/applications
3. Click on "New Application"
4. Fill in the information:
   - Name: Export Trakt 4 Letterboxd
   - Redirect URL: urn:ietf:wg:oauth:2.0:oob
   - Description: (optional)
5. Save the application
6. Note your Client ID and Client Secret

### 2. Set up the configuration file

Copy the example configuration file to create your own:

```bash
cp .config.cfg.example .config.cfg
```

You can edit the configuration file manually if you prefer, but it's recommended to use the setup script in the next step.

### 3. Authentication configuration

Run the configuration script:

```bash
./setup_trakt.sh
```

This script will guide you through the following steps:

1. Enter your Client ID and Client Secret
2. Obtain an authorization code
3. Generate access tokens

## Usage

### Export your data

```bash
./Export_Trakt_4_Letterboxd.sh [option]
```

Available options:

- `normal` (default): Exports rated movies, rated episodes, movie and TV show history, and watchlist
- `initial`: Exports only rated and watched movies
- `complet`: Exports all available data

### Result

The script generates a `letterboxd_import.csv` file that you can import on Letterboxd at the following address: https://letterboxd.com/import/

## Troubleshooting

### No data is exported

If the script runs without error but no data is exported:

1. Check that your Trakt.tv profile is public
2. Verify that you have correctly configured authentication
3. Run the configuration script again: `./setup_trakt.sh`

### Authentication errors

If you encounter authentication errors:

1. Check that your Client ID and Client Secret are correct
2. Get a new access token by running `./setup_trakt.sh`

## License

This project is under MIT license.

## Authors

👤 **u2pitchjami**

- Twitter: [@u2pitchjami](https://twitter.com/u2pitchjami)
- Github: [@u2pitchjami](https://github.com/u2pitchjami)
- LinkedIn: [@thierry-beugnet-a7761672](https://linkedin.com/in/thierry-beugnet-a7761672)

## Documentation

thanks to :

https://gist.github.com/kijart/4974b7b61bcec092dc3de3433e6e00e2

https://gist.github.com/darekkay/ff1c5aadf31588f11078
