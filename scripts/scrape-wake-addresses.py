#!/usr/bin/env python3
"""
Scrape all unique store addresses from Wake County ABC website
"""
import json
import urllib.request
import urllib.parse
from html.parser import HTMLParser

class AddressParser(HTMLParser):
    def __init__(self):
        super().__init__()
        self.addresses = set()
        self.in_address_span = False
        self.current_address = ""

    def handle_starttag(self, tag, attrs):
        if tag == "span":
            for attr, value in attrs:
                if attr == "class" and value == "address":
                    self.in_address_span = True

    def handle_endtag(self, tag):
        if tag == "span" and self.in_address_span:
            self.in_address_span = False
            if self.current_address:
                # Clean up address (remove <br/> tags)
                addr = self.current_address.replace("<br/>", " ").replace("<br>", " ")
                addr = " ".join(addr.split())  # Normalize whitespace
                self.addresses.add(addr)
                self.current_address = ""

    def handle_data(self, data):
        if self.in_address_span:
            self.current_address += data

def fetch_addresses(product_name):
    """Fetch addresses for a given product search"""
    url = "https://wakeabc.com/search-results"
    data = urllib.parse.urlencode({"productSearch": product_name}).encode()
    headers = {
        "User-Agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36",
        "Content-Type": "application/x-www-form-urlencoded",
        "Referer": "https://wakeabc.com/search-our-inventory/"
    }

    req = urllib.request.Request(url, data=data, headers=headers)
    with urllib.request.urlopen(req) as response:
        html = response.read().decode()
        parser = AddressParser()
        parser.feed(html)
        return parser.addresses

def main():
    # Search multiple products to get all store addresses
    search_terms = ["Buffalo Trace", "Weller", "Eagle Rare"]
    all_addresses = set()

    for term in search_terms:
        addresses = fetch_addresses(term)
        all_addresses.update(addresses)

    # Convert to sorted list
    addresses_list = sorted(all_addresses)

    # Output as JSON
    print(json.dumps(addresses_list, indent=2))

if __name__ == "__main__":
    main()
