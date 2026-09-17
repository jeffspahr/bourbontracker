const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const test = require('node:test');
const vm = require('node:vm');

// Exercise the shipped inline application without a browser or third-party packages.
const html = fs.readFileSync(path.join(__dirname, '..', 'index.html'), 'utf8');
const script = [...html.matchAll(/<script\b[^>]*>([\s\S]*?)<\/script>/g)]
    .map(match => match[1]).find(source => source.includes('function loadInventory'));
assert.ok(script, 'The page contains its inventory application script');

class Element {
    constructor(tagName = 'div') {
        this.tagName = tagName.toUpperCase();
        this.style = {};
        this.dataset = {};
        this.attributes = new Map();
        this.children = [];
        this.value = '';
        this.hidden = false;
        this.className = '';
        this.classList = { add() {}, remove() {}, toggle() {} };
        this.textContent = '';
    }
    set innerHTML(value) { this.children = []; this._html = value; }
    get innerHTML() { return this._html || ''; }
    appendChild(child) { this.children.push(child); return child; }
    append(...children) { this.children.push(...children); }
    replaceChildren(...children) { this.children = children; }
    setAttribute(name, value) { this.attributes.set(name, String(value)); }
    getAttribute(name) { return this.attributes.get(name) ?? null; }
    hasAttribute(name) { return this.attributes.has(name); }
    removeAttribute(name) { this.attributes.delete(name); }
    addEventListener() {}
    contains(target) { return target === this || this.children.some(child => child.contains?.(target)); }
    focus() {}
    querySelector(selector) {
        return this.querySelectorAll(selector)[0] || null;
    }
    querySelectorAll(selector) {
        const matches = child => selector.startsWith('.')
            ? child.className.split(' ').includes(selector.slice(1))
            : selector === 'input[type="checkbox"]' && child.type === 'checkbox';
        return this.children.flatMap(child => [
            ...(matches(child) ? [child] : []),
            ...(child.querySelectorAll?.(selector) || [])
        ]);
    }
}

function inventoryItem(state, productId, storeId = '100') {
    return {
        'bt.state': state,
        'bt.productId': productId,
        'bt.productName': `${state} product ${productId}`,
        'bt.storeId': storeId,
        'bt.quantity': 3,
        'bt.listingType': 'Allocation',
        'geo.location': { lat: state === 'VA' ? 37.5 : 35.8, lon: -78.6 },
        '@timestamp': '2026-09-17T12:00:00Z'
    };
}

function response(inventory) {
    return { ok: true, json: async () => inventory };
}

function deferred() {
    let resolve;
    const promise = new Promise(done => { resolve = done; });
    return { promise, resolve };
}

function application(fetch) {
    const elements = new Map();
    const getElement = id => {
        if (!elements.has(id)) elements.set(id, new Element());
        return elements.get(id);
    };
    getElement('error').hidden = true;
    getElement('product-dropdown').style.display = 'none';
    const map = { fitBounds() {}, setZoom() {}, getZoom: () => 7 };
    const context = vm.createContext({
        console: { log() {}, error() {}, warn() {} },
        URLSearchParams,
        AbortController,
        setTimeout,
        clearTimeout,
        fetch,
        localStorage: { getItem: () => null, setItem() {} },
        document: {
            getElementById: getElement,
            createElement: tag => new Element(tag),
            querySelectorAll: selector => selector === '.filter-buttons button'
                ? [getElement('select-products-button'), getElement('clear-products-button')]
                : [],
            addEventListener() {},
            body: new Element('body'),
            head: new Element('head')
        },
        window: { location: { search: '' }, matchMedia: () => ({ matches: false }) },
        google: { maps: {
            Marker: class {
                constructor(options) { Object.assign(this, options); }
                setMap(value) { this.map = value; }
                getPosition() { return this.position; }
                addListener() {}
            },
            LatLngBounds: class { extend() {} },
            event: { addListenerOnce() {} }
        } },
        testMap: map
    });
    vm.runInContext(script, context, { filename: 'index.html' });
    const run = source => vm.runInContext(source, context);
    run("regionInitialized = true; selectedRegion = 'VA'; map = testMap; infoWindow = { close() {} };");
    return { run, getElement, context };
}

test('changing region removes unavailable product selections and refreshes chips', async () => {
    const app = application(async url => response([
        inventoryItem(url.includes('-va.') ? 'VA' : 'NC', url.includes('-va.') ? 'va-only' : 'nc-only')
    ]));
    await app.run('loadInventory()');
    app.run("toggleProduct('va-only')");
    assert.equal(app.run('selectedProducts.size'), 1);
    app.getElement('region-select').value = 'NC';
    await app.run('filterByRegion()');
    assert.equal(app.run('selectedProducts.size'), 0);
    assert.equal(app.getElement('selected-products').children.length, 0);
    assert.equal(app.run("allInventory[0]['bt.state']"), 'NC');
    assert.equal(String(app.getElement('stores-count').textContent), '1');
    assert.doesNotThrow(() => app.run("toggleProduct('nc-only')"));
});

test('late inventory responses cannot replace the most recently selected region', async () => {
    const started = deferred();
    const older = deferred();
    const app = application(url => {
        if (url.includes('-va.')) {
            started.resolve();
            return older.promise;
        }
        return Promise.resolve(response([inventoryItem('NC', 'nc-only')]));
    });
    const firstLoad = app.run('loadInventory()');
    await started.promise;
    app.getElement('region-select').value = 'NC';
    await app.run('filterByRegion()');
    older.resolve(response([inventoryItem('VA', 'va-only')]));
    await firstLoad;
    assert.equal(app.run('selectedRegion'), 'NC');
    assert.equal(app.run("allInventory[0]['bt.state']"), 'NC');
    assert.equal(app.run("Object.hasOwn(allProducts, 'va-only')"), false);
    assert.equal(app.getElement('loading').style.display, 'none');
});

test('failed loads remove stale markers and a later successful load clears the error', async () => {
    let fail = false;
    const app = application(async () => {
        if (fail) throw new Error('Simulated unavailable inventory');
        return response([inventoryItem('VA', 'va-only')]);
    });
    await app.run('loadInventory()');
    app.run("toggleProduct('va-only')");
    assert.equal(app.getElement('selected-products').children.length, 1);
    const oldMarker = app.run('markers[0]');
    assert.ok(oldMarker);
    fail = true;
    await app.run('loadInventory()');
    assert.equal(app.getElement('error').hidden, false);
    assert.equal(app.run('markers.length'), 0);
    assert.equal(oldMarker.map, null);
    const inventoryControls = [
        'product-search', 'listing-select', 'select-products-button', 'clear-products-button'
    ];
    for (const id of inventoryControls) {
        assert.equal(app.getElement(id).disabled, true, `${id} cannot act on stale inventory`);
    }
    assert.equal(app.getElement('selected-products').hidden, true);
    fail = false;
    await app.run('loadInventory()');
    assert.equal(app.getElement('error').hidden, true);
    assert.equal(app.getElement('loading').style.display, 'none');
    assert.equal(app.run('markers.length'), 1);
    for (const id of inventoryControls) {
        assert.equal(app.getElement(id).disabled, false, `${id} is usable after recovery`);
    }
    assert.equal(app.getElement('selected-products').hidden, false);
    assert.equal(app.getElement('selected-products').children.length, 1);
});

test('empty filtered results retain inventory freshness without displaying Invalid Date', () => {
    const app = application(async () => response([]));
    app.context.fixture = [inventoryItem('VA', 'va-only')];
    app.run('allInventory = fixture; updateStats(allInventory)');
    const freshness = String(app.getElement('last-updated').textContent);
    app.run('updateStats([])');
    assert.equal(String(app.getElement('total-items').textContent), '0');
    assert.equal(String(app.getElement('last-updated').textContent), freshness);
    assert.doesNotMatch(freshness, /Invalid Date/);
    app.run('allInventory = []; updateStats([])');
    assert.doesNotMatch(String(app.getElement('last-updated').textContent), /Invalid Date/);
});

test('stores with identical IDs in different states count separately', () => {
    const app = application(async () => response([]));
    app.context.fixture = [
        inventoryItem('VA', 'va-only', '100'),
        inventoryItem('NC', 'nc-only', '100'),
        inventoryItem('VA', 'va-other', '100')
    ];
    app.run('allInventory = fixture; updateStats(allInventory)');
    assert.equal(String(app.getElement('stores-count').textContent), '2');
});
