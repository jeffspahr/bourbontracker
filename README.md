# VA ABC Bourbon Tracker
This is intended to run in Kubernetes, but it can be modified to fit your environment.  The app will output inventory data in JSON format which will be picked up by Filebeat and shipped to an Elasticsearch cluster.  I plan to include the end to end manifests to deploy this in any Kubnernetes cluster.

The end goal is to be able to visualize and alert on geolocation inventory data.

Inspired by https://github.com/misfitlabs/pappytracker

# Convert Product List from a Python Dictionary to JSON
```python3 dict2json.py |jq > products.json```
Tested on python3.

# Generate List of Stores
```go run generateStoreList.go```

# Run the Tracker
```go run tracker.go```
