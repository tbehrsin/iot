
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
                transaction.reject(new Error(response.error));
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
    BleTable.set(this, { device });
  }

  async connect() {
    const { device } = BleTable.get(this);
    d = await ble.connectToDevice(device.id, { autoConnect: true, requestMTU: 23 });
    await d.discoverAllServicesAndCharacteristics();
    return new BleConnection(d);
  }
}

class BleManager {
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
              resolve(new BleDevice(device));
            }
          );
        }
      }, true);
    });
  }

  stop() {
    ble.stopDeviceScan();
  }
}

export default new BleManager();
