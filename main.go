package main

import(
  "context"
  "fmt"
  "os"
  "github.com/akamensky/argparse"
  "github.com/hekmon/transmissionrpc/v3"
  "github.com/abriel/transmission-rutracker-seed-go/lib/config"
  "github.com/abriel/transmission-rutracker-seed-go/lib/rutracker"
  "github.com/abriel/transmission-rutracker-seed-go/lib/torrent"
  "github.com/abriel/transmission-rutracker-seed-go/lib/version"
)

func main() {
  parser := argparse.NewParser(os.Args[0], "Checks updates on rutracker.org for torrents in your BT client")
  run_configuration := parser.Flag("c", "configure", &argparse.Options{Required: false, Help: "Run configuration menu"})
  print_version := parser.Flag("v", "version", &argparse.Options{Required: false, Help: "Print the version and exit"})
  err := parser.Parse(os.Args)
  if err != nil {
    fmt.Print(parser.Usage(err))
  }

  if *print_version {
    fmt.Println("Version:", version.Version)
    return
  }

  if *run_configuration {
    config.Run()
    return
  }

  myconfig, err := config.New()
  if err != nil {
    fmt.Printf("Cannot parse or validate configuration file. Either fix or delete it.\nDetails: %s\n", err)
    return
  }

  transmission_client, _ := transmissionrpc.New(myconfig.BtUri(), nil)

  torrents, err := transmission_client.TorrentGetAll(context.TODO())
  if err != nil {
    fmt.Printf("Cannot get list of torrents. Check 'bt_uri' in configuration.\nDetails: %s\n", err)
    return
  }

  for _, torrent := range torrents {
    torrent_helpers := torrentHelpers.New(&torrent)

    rutracker_topic_url, err := torrent_helpers.GetRutrackerTopicUrl()
    if err != nil {
      fmt.Printf("Skipping\t%s\n", *torrent.Name)
      continue
    }

    fmt.Printf("Checking\t[%s]\t%s\n", rutracker_topic_url.String(), *torrent.Name)
    rutracker_page, err := rutracker.Page(rutracker_topic_url, myconfig.UserAgent)
    if err != nil {
      fmt.Printf("Failed to fetch the web page; %v\n", err)
      continue
    }

    new_torrent_hash, err := rutracker_page.GetTorrentHash()
    if err != nil {
      fmt.Printf("Error during processing the web page: %v\n", err)
      continue
    }

    current_torrent_hash := torrent_helpers.GetTorrentHash()

    if new_torrent_hash != current_torrent_hash {
      fmt.Printf("Found new version %s for [%s] %s\n", new_torrent_hash, current_torrent_hash, *torrent.Name)

      magnet_url, _ := rutracker_page.GetMagnetUrl()
      warns, err := torrentHelpers.ReplaceTorrent(transmission_client, &torrent, magnet_url, rutracker_topic_url.String())
      if err != nil {
        fmt.Printf("Error occurred while replacing the torrent: %v\n", err)
        continue
      }

      fmt.Printf("Torrent [%v] has been updated\n", *torrent.Name)

      if len(warns) > 0 {
        fmt.Printf("However there are some warnings:\n")

        for i, warning := range(warns) {
          fmt.Printf("%d. %s\n", i, warning)
        }
      }
    }
  }
}
