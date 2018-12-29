
import { Client, Message } from 'react-native-paho-mqtt';

const WebSocketTable = new WeakMap();

const storage = {
  setItem: (key, item) => {
    storage[key] = item;
  },
  getItem: (key) => storage[key],
  removeItem: (key) => {
    delete storage[key];
  },
};

class WebSocketService {
  constructor() {
    WebSocketTable.set(this, { client: null, subscriptions: {}, buffer: [] });
  }

  async connect(uri) {
    const client = new Client({ uri, clientId: 'iot-mobile-' + (0x100000 + Math.random() * 0x7fffff).toString(), storage });

    client.on('connectionLost', (responseObject) => {
      if (responseObject.errorCode !== 0) {
        console.log(responseObject.errorMessage);
      }
    });

    client.on('messageReceived', (message) => {
      const { subscriptions, buffer } = WebSocketTable.get(this);
      const { destinationName: topic, payloadString: payload } = message;
      const body = JSON.parse(payload);

      for (const [k, v] of Object.entries(subscriptions)) {
        const re = new RegExp(`^${k.replace(/[-[\]{}()*?.,\\^$|\s]/g, '\\$&').replace(/#/g, '(.*)').replace(/\+/g, '([^/]+)')}$`);
        
        let match;
        if (match = topic.match(re)) {
          for (const subscription of v) {
            subscription.handler(topic, body, match);
          }
        }
      }
    });

    try {
      await client.connect({ useSSL: true })
      Object.assign(WebSocketTable.get(this), { client });

      const { subscriptions, buffer } = WebSocketTable.get(this);

      for (const v in subscriptions) {
        client.subscribe(v);
      }

      for (const { topic, object, qos } of buffer.splice(0, buffer.length)) {
        client.send(Object.assign(new Message(JSON.stringify(object)), { destinationName: topic, qos }));
      }
    } catch (error) {
      throw new ResourceError({ message: error.errorMessage || error.message, code: ResourceError.Forbidden });
    }
  }

  disconnect() {
    const { client } = WebSocketTable.get(this);
    client.disconnect();
  }

  publish(topic, object = null, qos = 2) {
    const { client, buffer } = WebSocketTable.get(this);

    if (client) {
      client.send(Object.assign(new Paho.MQTT.Message(JSON.stringify(object)), { destinationName: topic, qos }));
    } else {
      buffer.push({ topic, object, qos });
    }
  }

  subscribe(topic, handler, qos = 2) {
    if (typeof topic !== 'string') {
      throw new TypeError('topic must be a string');
    }

    if (typeof handler !== 'function') {
      throw new TypeError('handler must be a function');
    }

    if ([0, 1, 2].indexOf(qos) === -1) {
      throw new TypeError('qos must be 0, 1 or 2');
    }

    const { client, subscriptions } = WebSocketTable.get(this);

    const subscription = {
      handler
    };

    if (!subscriptions[topic]) {
      subscriptions[topic] = [];

      if (client) {
        client.subscribe(topic);
      }
    }

    subscriptions[topic].push(subscription);

    return {
      unsubscribe: () => {
        const index = subscriptions[topic].indexOf(subscription);

        if (index === -1) {
          return;
        }

        subscriptions[topic].splice(index, 1);

        if (subscriptions[topic].length === 0) {
          delete subscriptions[topic];
          if (client) {
            client.unsubscribe(topic);
          }
        }

        console.info(index, subscriptions);
      }
    }
  }
}

export default () => new WebSocketService();
