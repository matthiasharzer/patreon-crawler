# Patreon Crawler
A simple tool to download all media from a patreon creator

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
<br>

## Setup
Download the [latest release](https://github.com/MatthiasHarzer/patreon-crawler/releases) and add the executable to your `PATH`.

> Don't forget to make the file executable using `chmod +x patreon-crawler` when using a UNIX-based operating system

## Usage

The `patreon-crawler` requires you to manually extract the cookie from the Patreon website inorder to authenticate you against the Patreon API. 

### Extracting the cookies from Patreon
The simplest way to extract the cookies is by visiting the Patreon home page and copying the cookie-value from a network request:

1. In a new tab, [open your browser's network tab](https://superuser.com/questions/1625410/is-there-a-shortcut-for-opening-the-dev-tools-network-tab-in-chrome).
2. Navigate to [patreon.com/home](https://www.patreon.com/home). Make sure you are logged in.
3. Find a request starting with `current_user?include=....`. Click it a scroll to the `Request Headers` section.
4. Copy the `Cookie` value to your clipboard.

> To authenticate against `patreon.com` it's actually sufficient to copy the `session_id=<id>` cookie, however, passing the entire cookie string won't affect functionality.

### Using the tool

To download all media from a creator, run the following command.

```shell
patreon-crawler --creator <creator-id>
```

You will be prompted to enter the cookie (the one you copied earlier) and a download directory. 

If you do not wish do be prompted, you can also use the `--cookie` and `--download-dir` flag respectively.

### Command line flags

The `patreon-crawler` supports the following command line flags.

| Argument                        | Description                                                                                                                                                                           |
|---------------------------------|---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `--creator <creator-id>`        | The creator ID to download media from. You can find this in the URL when visiting a creators page: `patreon.com/c/<creator-id>/...`                                                   |
| `--cookie <cookie-string>`      | The cookie from the Patreon website to authenticate against the Patreon API                                                                                                           |
| `--download-dir <directory>`    | The base directory to download media to. All files will be located in `<download-dir>/<creator>`                                                                                      |
| `--download-limit <number>`     | The maximum number of media files to download.                                                                                                                                        |
| `--download-inaccessible-media` | Whether to download media that is inaccessible (blurred images)                                                                                                                       |
| `--grouping <none \| by-post>`  | The strategy for grouping post media into folders. <br>`none` - Puts all media into the same folder (per creator)<br>`by-post` - Creates a folder for each post, containing its media |

