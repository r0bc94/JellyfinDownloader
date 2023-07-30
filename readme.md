# Jellyfin Downloader

This is a simple tool written in Go to download a Series or a single season from a given jellyfin instance. 

## Disclaimer

Since I've written this in order to learn go, not everything might be very written _as it should_. Also keep in mind
that not everything might be work as expected. 

## Usage

Currently, you have to provide the series id or also the season id. You can extract the ID from the jellyfin, as its a simple url parameter. 

For example, if you click on a series you want to download, the series ID can be found here: 

```
http://localhost:8096/web/index.html#!/details?

id=e9c503bbdb25e07b71e0168298402e18 <-- This is the ID you need to copy

&context=tvshows&serverId=da596f62e19b4ee296431dc373bad050
```

To download all episodes of a given series, call the tool like so: 

```bash
jellyfindownloader \
    -url <BaseURL of the JF Server> \
    -seriesid <ID of the series you want to download> 
```

If you only want to download a specific season, you have to find out the season Id first. You can also find the season ID in the URL when clicking on the season you want to download: 

```
http://localhost:8096/web/index.html#!/details?

id=5a0e9809ef72219c159e130235e9940d <-- This is the ID you need to copy&

serverId=da596f62e19b4ee296431dc373bad050
```

To download a specific episode, you need to call the tool like this: 

```bash
jellyfindownloader \
    -url <BaseURL of the JF Server> \
    -seriesid <ID of the series you want to download> 
    -seasonid <ID of the season to download>
```

You can also pass additional argument such as the username or password. If those are passed, you do not need to provide them when running the script. Use `-h` for more information: 

```
./jellyfindownloader -h
Usage of ./jellyfindownloader:
  -password string
        Passwort for the Jellyfin instance. If not provided, username will be prompted.
  -seasonid string
        If given, only the episodes with the provided season Id will be downloaded
  -seriesid string
        ID which points to the series which should be downloaded
  -url string
        Base URL which points to the Jellyfin Instance
  -username string
        Username used to login to the Jellyfin instance. If not provided, password will be prompted.
```

## Todo

- [ ] Instead of fiddling with Ids, one should only provide the series name and episode number which should be downloaded
- [ ] Retrieve the credentials via an environnement variable
- There are probably some other things which should be done whose will come to my mind later on


