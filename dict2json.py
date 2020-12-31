import json
import products as const

# We might as well invert the dict/map here so we don't have to do it later.
# This is important so we can look up the product name (value) by its ID (key).
# We will never need to look the product ID by the product name so there's also no need to maintain two maps.
invertList = {v: k for k, v in const.products.items()}

print (json.dumps(invertList))