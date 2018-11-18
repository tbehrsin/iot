
const BleTable = new WeakMap();

const sleep = (ms) => new Promise((resolve) => setTimeout(resolve, ms));

class BleConnection {
  async send(json) {
    const response = await fetch('http://iot-gateway.local/blemu', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify(json)
    });
    const jso = await response.json();

    if(jso.error) {
      throw new ResourceError(jso.error);
    }

    return jso;
  }
}

class BleDevice {
  constructor() {
    BleTable.set(this, { connection: null });
  }

  async connect() {
    const table = BleTable.get(this);

    if (table.connection) {
      throw new Error('already connected');
    }

    await sleep(2000);
    table.connection = new BleConnection();
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

  async start() {
    const table = BleTable.get(this);

    await sleep(2000);
    table.device = new BleDevice();
    this.stop();
    return table.device;
  }

  stop() {

  }

  get device() {
    const { device } = BleTable.get(this);
    return device;
  }

  get connection() {
    return this.device && this.device.connection;
  }
}

export default BleService;
