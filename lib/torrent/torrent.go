package torrentHelpers

import(
  "context"
  "fmt"
  gourl "net/url"
  "github.com/hekmon/transmissionrpc/v3"
  "strings"
)

func GetAnnotationUrl(magnet_url string, pk string) (url string, err error) {
  magnet_gourl, err := gourl.Parse(magnet_url)
  if err != nil {
    return
  }

  trackers := magnet_gourl.Query()["tr"]
  if len(trackers) == 0 {
    err = fmt.Errorf("Could not find 'tr' in magnet url %#v", magnet_url)
    return
  }

  tracker_url, err := gourl.Parse(trackers[0])
  if err != nil {
    return
  }

  tracker_url.RawQuery = gourl.Values{"pk": []string{pk}}.Encode()
  url = tracker_url.String()

  return
}

func ReplicateAnnotationUrl(client *transmissionrpc.Client, new_torrent *transmissionrpc.Torrent, torrent *transmissionrpc.Torrent) (err error) {
  pk, err := New(torrent).GetPk()
  if err != nil {
    return
  }

  ann_url, err := GetAnnotationUrl(*new_torrent.MagnetLink, pk)
  if err != nil {
    return
  }

  err = client.TorrentSet(
    context.TODO(),
    transmissionrpc.TorrentSetPayload{
      TrackerList: []string{ann_url},
      IDs: []int64{*new_torrent.ID},
    },
  )

  return
}

func ReplaceTorrent(client *transmissionrpc.Client, torrent *transmissionrpc.Torrent, magnet_url string, label string) (warnings []error ,err error) {
  paused := true

  new_torrent, err := client.TorrentAdd(
    context.TODO(),
    transmissionrpc.TorrentAddPayload{
      DownloadDir: torrent.DownloadDir,
      Labels: []string{label},
      Paused: &paused,
      PeerLimit: torrent.PeerLimit,
      BandwidthPriority: torrent.BandwidthPriority,
      Filename: &magnet_url,
    },
  )
  if err != nil {
    return
  }

  new_torrents, err := client.TorrentGetAllFor(context.TODO(), []int64{*new_torrent.ID})
  if err != nil {
    return
  }
  if len(new_torrents) == 0 {
    err = fmt.Errorf("Could not fetch just added torrent with id %d", *new_torrent.ID)
    return
  }
  new_torrent = new_torrents[0]

  will_start := *torrent.Status != transmissionrpc.TorrentStatusStopped

  err = ReplicateAnnotationUrl(client, &new_torrent, torrent)
  if err != nil {
    warnings = append(warnings, fmt.Errorf("BT client won't be able to send seeding stats to a tracker due to: %v", err))
  }

  err = client.TorrentRemove(
    context.TODO(),
    transmissionrpc.TorrentRemovePayload{
      IDs: []int64{*torrent.ID},
      DeleteLocalData: false,
    },
  )
  if err != nil {
    err = fmt.Errorf("Failed to remove old torrent with ID %v. Will not start new torrent %v to avoid conflicts", *torrent.ID, *new_torrent.ID)
    return
  }

  if will_start {
    err = client.TorrentStartIDs(context.TODO(), []int64{*new_torrent.ID})
  }

  return
}

type Torrent struct {
  torrent *transmissionrpc.Torrent
}

func New(torrent *transmissionrpc.Torrent) *Torrent {
  return &Torrent{ torrent: torrent }
}

func (h *Torrent) GetRutrackerTopicUrl() (url *gourl.URL, err error) {
  url, err = gourl.Parse(*h.torrent.Comment)
  if err == nil && url.Host != "" {
    return
  }

  for _, label := range h.torrent.Labels {
    url, err = gourl.Parse(label)
    if err == nil && url.Host != "" {
      return
    }
  }

  err = fmt.Errorf("Neither Comment nor Labels have a URL")
  return
}

func (h *Torrent) GetTorrentHash() string {
  return strings.ToUpper(*h.torrent.HashString)
}

func (h *Torrent) GetPk() (string, error) {
  for _, tracker := range(h.torrent.Trackers) {
    url, err := gourl.Parse(tracker.Announce)
    if err != nil {
      continue
    }

    for k, v := range(url.Query()) {
      if k == "pk" {
        return v[0], nil
      }
    }
  }

  return "", fmt.Errorf("Fail to find pk in %#v", h.torrent.Trackers)
}
