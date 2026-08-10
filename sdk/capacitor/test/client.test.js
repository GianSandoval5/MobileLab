import assert from 'node:assert/strict';
import test from 'node:test';

import {
  MobileLabClient,
  MobileLabLifecycleReporter,
  MOBILELAB_PROTOCOL_VERSION,
} from '../dist/index.js';

class RecordingTransport {
  events = [];

  async send(event) {
    this.events.push(event);
  }
}

test('client emits the versioned Capacitor contract', async () => {
  const transport = new RecordingTransport();
  const client = new MobileLabClient({
    endpoint: 'http://10.0.2.2:4566',
    sessionId: 'run-1',
    transport,
  });

  await client.marker('checkout.loaded', { screen: 'checkout' });
  await client.assertThat('cart.total', true);

  assert.equal(transport.events.length, 2);
  assert.equal(transport.events[0].protocolVersion, MOBILELAB_PROTOCOL_VERSION);
  assert.equal(transport.events[0].framework, 'capacitor');
  assert.equal(transport.events[0].sessionId, 'run-1');
  assert.equal(transport.events[1].passed, true);
});

test('lifecycle reporter registers asynchronously and deduplicates state', async () => {
  const transport = new RecordingTransport();
  const client = new MobileLabClient({ endpoint: 'http://127.0.0.1:4566', transport });
  let listener;
  let removed = false;
  const app = {
    async addListener(eventName, callback) {
      assert.equal(eventName, 'appStateChange');
      listener = callback;
      return { async remove() { removed = true; } };
    },
  };
  const reporter = new MobileLabLifecycleReporter(client, app);

  await reporter.attach(false);
  listener({ isActive: false });
  listener({ isActive: false });
  listener({ isActive: true });
  await Promise.resolve();
  await reporter.detach();

  assert.deepEqual(transport.events.map((event) => event.name), ['background', 'foreground']);
  assert.equal(removed, true);
});

test('HTTP transport posts JSON to the SDK bridge path', async () => {
  let requestUrl;
  let requestInit;
  const client = new MobileLabClient({
    endpoint: 'http://127.0.0.1:4566/nested/path',
    fetchImplementation: async (url, init) => {
      requestUrl = url;
      requestInit = init;
      return new Response(null, { status: 202 });
    },
  });

  await client.marker('transport.ready');

  assert.equal(requestUrl.pathname, '/__mobilelab/sdk/events');
  assert.equal(requestInit.headers['X-MobileLab-SDK'], '@mobilelab/capacitor/1.1.0');
  assert.equal(JSON.parse(requestInit.body).name, 'transport.ready');
});
