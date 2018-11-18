
import BleServiceFactory from './ble';
import WebSocketServiceFactory from './ws';

const factories = {
  ble: BleServiceFactory,
  ws: WebSocketServiceFactory
};

const services = {
  initialize: (initialProps) => {
    for (const [name, factory] of Object.entries(factories)) {
      services[name] = new factory(initialProps);
    }
  }
};
module.exports = services;
