
import React from 'react';
import { BleManager as NativeBleManager } from 'react-native-ble-plx';
import { constants } from '../../../../app.json';
import { Buffer } from 'buffer';

const ble = new NativeBleManager();
const BleTable = new WeakMap();

class BleConnection {
  constructor(device) {
    BleTable.set(this, {
      device,
      transactions: { nextId: 0 }
    });

    const subscription = device.monitorCharacteristicForService(
      constants.bleService,
      constants.bleCharacteristic,
      async (err) => {
        if(err) {
          console.error(err);
          return;
        }

        try {
          while(true) {
            const char = await device.readCharacteristicForService(
              constants.bleService,
              constants.bleCharacteristic
            );

            const buffer = Buffer(char.value, 'base64');

            if (buffer.length === 0) {
              return;
            }

            const id = buffer.readUInt8(0);
            const data = buffer.slice(1);

            const { transactions } = BleTable.get(this);
            const transaction = transactions[id];

            if (data.length > 0) {
              transaction.response.push(data);
            } else {
              const response = JSON.parse(Buffer.concat(transaction.response).toString());
              delete transactions[id];

              if (response.error) {
                transaction.reject(new ResourceError(response.error));
              } else {
                transaction.resolve(response);
              }
            }
          }
        } catch(err) {
          console.error(err);
          return;
        }
      }
    );
  }

  async send({ type, payload }) {
    const { device, transactions } = BleTable.get(this);

    const id = transactions.nextId++;
    transactions.nextId %= 256;
    const transaction = {
      response: []
    };
    transactions[id] = transaction;

    const p = new Promise((res, rej) => {
      transaction.resolve = res;
      transaction.reject = rej;
    });

    const data = Buffer.from(JSON.stringify({ type, payload }));
    for (let i = 0; i < data.length; i += 19) {
      const buffer = Buffer.concat([
        Buffer.from([id]),
        Buffer.from(data.slice(i, i + 19))
      ]).toString('base64');
      await device.writeCharacteristicWithoutResponseForService(
        constants.bleService,
        constants.bleCharacteristic,
        buffer
      );
    }
    const buffer = Buffer.from([id]).toString('base64');
    await device.writeCharacteristicWithoutResponseForService(
      constants.bleService,
      constants.bleCharacteristic,
      buffer
    );

    return p;
  }
}

class BleDevice {
  constructor(device) {
    BleTable.set(this, { device, connection: null });
  }

  async connect() {
    const table = BleTable.get(this);

    if (table.connection) {
      throw new Error('already connected');
    }
    d = await ble.connectToDevice(table.device.id, { autoConnect: true, requestMTU: 23 });
    await d.discoverAllServicesAndCharacteristics();
    table.connection = new BleConnection(d);
    return table.connection;
  }

  get connection() {
    const { connection } = BleTable.get(this);
    return connection;
  }
}

class BleService {
  constructor() {
    BleTable.set(this, {
      device: null
    });
  }

  start() {
    return new Promise((resolve, reject) => {
      ble.onStateChange((newState) => {
        if (newState === 'PoweredOn') {
          ble.startDeviceScan(
            [constants.bleService],
            {},
            (err, device) => {
              if(err) {
                reject(err);
                return;
              }

              const table = BleTable.get(this);

              if (table.device) {
                return;
              }

              table.device = new BleDevice(device);
              this.stop();
              resolve(table.device);
            }
          );
        }
      }, true);
    });
  }

  stop() {
    ble.stopDeviceScan();
  }

  get device() {
    const { device } = BleTable.get(this);
    return device;
  }

  get connection() {
    return this.device && this.device.connection;
  }
}

export default ({ isEmulator = false }) => {
  if (isEmulator) {
    const BleService = require('./index.dev').default;
    return new BleService();
  }

  return new BleService();
};
