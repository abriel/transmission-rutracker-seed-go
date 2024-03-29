package rutracker

import(
  "fmt"
  "io"
  "regexp"
  gourl "net/url"
  "net/http"
)

type page struct {
  content string
  magnet_url string
  hash string
}

func Page(url *gourl.URL, userAgent string) (*page, error) {
  request := http.Request{
    URL: url,
    Header: http.Header{
      http.CanonicalHeaderKey("User-Agent"): []string{userAgent},
    },
  }

  response, err := (&http.Client{}).Do(&request)
  if err != nil {
    return nil, err
  }
  defer response.Body.Close()

  if response.StatusCode != http.StatusOK {
    return nil, fmt.Errorf("Status code: %d, headers: %#v", response.StatusCode, response.Header)
  }

  bodyBytes, err := io.ReadAll(response.Body)
  if err != nil {
    return nil, err
  }

  return &page{content: string(bodyBytes)}, nil
}

func (p *page) GetMagnetUrl() (string, error) {
  if p.magnet_url != "" {
    return p.magnet_url, nil
  }

  regx, err := regexp.Compile(`magnet:\?xt=[^"]+`)
  if err != nil {
    panic(err)
  }

  p.magnet_url = regx.FindString(p.content)
  if p.magnet_url == "" {
    return "", fmt.Errorf("Magnet URL was not found")
  }

  return p.magnet_url, nil
}

func (p *page) GetTorrentHash() (string, error) {
  if p.hash != "" {
    return p.hash, nil
  }

  magnet_url, err := p.GetMagnetUrl()
  if err != nil {
    return "", err
  }

  regx, err := regexp.Compile(`btih:([0-9A-F]+)`)
  if err != nil {
    panic(err)
  }

  all_matches := regx.FindStringSubmatch(magnet_url)
  if len(all_matches) < 2 {
    return "", fmt.Errorf("Torrent hash was not found in a magnet URL %v", magnet_url)
  }
  p.hash = all_matches[1]

  return p.hash, nil
}
