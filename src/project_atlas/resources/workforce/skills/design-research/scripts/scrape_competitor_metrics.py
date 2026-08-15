import json
import urllib.request
from html.parser import HTMLParser

class SimpleMetricParser(HTMLParser):
    def __init__(self):
        super().__init__()
        self.in_title = False
        self.title = ""
        self.links = 0
        self.images = 0

    def handle_starttag(self, tag, attrs):
        if tag == "title":
            self.in_title = True
        elif tag == "a":
            self.links += 1
        elif tag == "img":
            self.images += 1

    def handle_data(self, data):
        if self.in_title:
            self.title += data

    def handle_endtag(self, tag):
        if tag == "title":
            self.in_title = False

def scrape_metrics(url):
    try:
        req = urllib.request.Request(url, headers={'User-Agent': 'Mozilla/5.0'})
        with urllib.request.urlopen(req) as response:
            html = response.read().decode('utf-8')
            
        parser = SimpleMetricParser()
        parser.feed(html)
        
        return {
            "url": url,
            "title": parser.title.strip(),
            "link_count": parser.links,
            "image_count": parser.images
        }
    except Exception as e:
        return {"url": url, "error": str(e)}

if __name__ == "__main__":
    urls = [
        "https://example.com",
        "https://www.w3.org"
    ]
    
    results = [scrape_metrics(url) for url in urls]
    print(json.dumps(results, indent=2))
