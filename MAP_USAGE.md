# Google Maps Visualization Setup

## Overview

The tracker has been modified to output inventory data to a JSON file that can be visualized on a Google Map instead of sending to Elasticsearch/Kibana.

## Changes Made

### 1. Modified tracker.go
- Now collects all inventory data in memory during execution
- Outputs to `inventory.json` file (formatted JSON array)
- Still only includes items with quantity > 0
- Prints summary: "Found X items in stock across all stores"

### 2. Created map.html
- Standalone HTML file with embedded Google Maps visualization
- Features:
  - Interactive map centered on Virginia
  - **Product filter dropdown** - Select specific bourbons to display on the map
  - Color-coded markers (green = high stock, orange = medium, red = low)
  - Click markers to see product details
  - Statistics dashboard showing total items, unique products, stores, and last update
  - Responsive design
  - Groups multiple products at the same store location
  - Real-time filtering with "Select All" and "Clear Filter" buttons

### 3. Updated Dockerfile
- Now includes map.html in the container image

## How to Use

### Step 1: Get a Google Maps API Key

1. Go to https://console.cloud.google.com/
2. Create or select a project
3. Enable the "Maps JavaScript API"
4. Create an API key under "APIs & Services" → "Credentials"
5. (Optional) Restrict the API key to your domain

### Step 2: Configure API Key (Secure Method)

Copy the example config file and add your API key:

```bash
cp config.example.js config.js
```

Then edit `config.js` and replace `YOUR_API_KEY_HERE` with your actual API key:

```javascript
const GOOGLE_MAPS_API_KEY = 'YOUR_ACTUAL_API_KEY';
```

**Important:** `config.js` is in `.gitignore` and will NOT be committed to version control, keeping your API key secure.

### Step 3: Run the Tracker

**Local execution:**
```bash
go build tracker.go
./tracker
```

This will create `inventory.json` in the current directory.

**Docker execution:**
```bash
docker build -t bourbontracker .
docker run --rm -v $(pwd):/root bourbontracker
```

The `-v` flag mounts the current directory so `inventory.json` is written to your host machine.

### Step 4: View the Map

The map needs to be served through a web server (browsers block loading local JSON files for security reasons).

**Local Development (Recommended):**

Start a simple web server in the project directory:

```bash
# Python 3
python3 -m http.server 8000

# Or Python 2
python -m SimpleHTTPServer 8000

# Or if you have Node.js
npx http-server -p 8000
```

Then open your browser to: **http://localhost:8000/map.html**

**Alternative - Direct File Access:**

Some browsers allow opening local files if you disable security:
- Chrome: `open -a "Google Chrome" --args --allow-file-access-from-files map.html`
- Firefox: May work directly without flags

**For Web Hosting:**
- Upload `map.html`, `config.js`, and `inventory.json` to the same directory on your web server
- Access via your domain (e.g., https://yourdomain.com/map.html)

## Automated Updates

### Option 1: Cron Job (Local)

Create a cron job to run the tracker periodically:

```bash
# Run every 15 minutes
*/15 * * * * cd /path/to/bourbontracker && ./tracker
```

### Option 2: Kubernetes CronJob

If you want to keep using Kubernetes but serve the map differently:

1. Mount a persistent volume to the pod
2. Write `inventory.json` to the volume
3. Serve the volume via a simple web server (nginx, caddy, etc.)

Example nginx configuration:
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: nginx-config
data:
  default.conf: |
    server {
      listen 80;
      root /data;
      location / {
        autoindex on;
      }
    }
```

### Option 3: GitHub Pages (Simple Static Hosting)

1. Create a GitHub repository
2. Push `map.html` and `inventory.json`
3. Enable GitHub Pages in repository settings
4. Set up GitHub Actions to run the tracker and commit updates

## Map Features

### Product Filter
Filter the map to show only specific bourbon products:
- **Dropdown menu** displays all available products sorted by total quantity
- Each product shows format: "Product Name (X bottles at Y locations)"
- **Multi-select** - Hold Ctrl/Cmd to select multiple products
- **Select All** button - View all products at once
- **Clear Filter** button - Reset to show all inventory
- Map and statistics update in real-time when filter changes
- Only shows stores that have the selected product(s)

Example use cases:
- Find all stores with Pappy Van Winkle
- Compare availability of Buffalo Trace vs Blanton's
- Focus on high-demand allocated items only

### Marker Colors
- **Green**: Store has > 10 items in stock
- **Orange**: Store has 6-10 items in stock
- **Red**: Store has 1-5 items in stock

### Info Windows
Click any marker to see:
- Store number
- All products in stock at that location
- Quantity for each product (color-coded)
- Product ID
- Direct link to ABC website for that product/store

### Statistics Bar
Shows real-time stats:
- Total items in stock across all stores
- Number of unique products available
- Number of stores with inventory
- Timestamp of last data update

## Differences from Elasticsearch/Kibana

### Advantages
- Simpler setup (just HTML + JSON file)
- No infrastructure needed (Elasticsearch, Filebeat, Kibana)
- Can host anywhere (S3, GitHub Pages, any web server)
- Faster load times for small datasets
- Better for snapshot/current state visualization

### Limitations
- No historical data (only shows latest run)
- No search/filter capabilities
- No time-series analysis
- No alerting built-in
- Manual refresh needed (reload page)

## Re-enabling Elasticsearch (Optional)

If you want both visualizations:

1. Keep the JSON file output
2. Add back the line-by-line stdout output:

```go
if pOut.Quantity > 0 {
    // For Elasticsearch
    pOutJSON, _ := json.Marshal(pOut)
    fmt.Println(string(pOutJSON))

    // For map visualization
    allInventory = append(allInventory, item)
}
```

This way Filebeat can still collect logs while the map uses the JSON file.

## Troubleshooting

### Map shows "Loading..." forever
- Check browser console for errors
- Verify `inventory.json` exists in same directory as `map.html`
- Confirm JSON file is valid (not empty, proper format)

### Map shows "Failed to load inventory data"
- **Most common cause**: Opening `map.html` directly as a file (file:// URL) instead of through a web server
  - **Solution**: Start a local web server: `python3 -m http.server 8000` and access via http://localhost:8000/map.html
- Ensure `inventory.json` is in the same directory as `map.html`
- Verify the tracker completed successfully
- Check browser console for CORS or fetch errors

### No markers appear on map
- Check if `inventory.json` has any items with quantity > 0
- Look for JavaScript errors in browser console
- Verify lat/lon coordinates are valid

### Map doesn't load at all
- Verify your Google Maps API key is valid
- Check if Maps JavaScript API is enabled in Google Cloud Console
- Look for API key errors in browser console

## Next Steps

Consider adding:
- Product filtering (show only specific bourbons)
- Search functionality
- Export to CSV
- Email alerts when specific products come in stock
- Historical comparison (store previous runs)
