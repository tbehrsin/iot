
const Color = require('color');
const XYConvert = require('@q42philips/hue-color-converter')
global.XYConvert = XYConvert;

global.controllers = [];

class HueAmbience extends zigbee.Controller {
  constructor() {
    super();
    global.controllers.push(this);

    console.info("HueAmbience", this.device);

    for (const ep of this.device.endpoints) {
      this.device.read(0x0000, 0x0004, ep.endpoint);
      this.device.read(0x0000, 0x0005, ep.endpoint);
    }

    this.color = '#ffffff';
  }

  set on(value) {
    this.setOn(value);
  }

  setOn(value, effect = null) {
    if (value) {
      this.device.send(0x0006, 0x01, 0x0b /*, new Uint8Array([]), { global: false } */);
    } else {
      this.device.send(0x0006, 0x00, 0x0b);
    }
  }

  set level(value) {
    this.setLevel(value);
  }

  setLevel(value, rate = 0) {
    const data = new ArrayBuffer(8);
    const view = new DataView(data);
    view.setUint8(0, value);
    view.setUint16(1, rate, true);
    this.device.send(0x0008, 0x00, 0x0b, data);
  }

  set color(value) {
    this.setRGB(value);
  }

  setRGB(value, rate = 0) {
    const { r, g, b } = Color(value).rgb().object();
    const [x, y] = XYConvert.calculateXY(r, g, b);
    return this.setXY({ x, y }, rate);
  }

  setXY({ x, y }, rate = 0) {
    const data = new ArrayBuffer(8);
    const view = new DataView(data);
    view.setUint16(0, Math.round(x * 0xffff), true);
    view.setUint16(2, Math.round(y * 0xffff), true);
    view.setUint16(4, rate, true);
    this.device.send(0x0300, 0x07, 0x0b, data);
  }

  onLeave() {

  }

  onUpdate() {

  }
}

HueAmbience.match = (device) => device.match({
  endpoints: [
    {
      id: 0x0b,
      clusters: [
        { id: 0x0000, type: 'in' },
        { id: 0x0003, type: 'in' },
        { id: 0x0004, type: 'in' },
        { id: 0x0005, type: 'in' },
        { id: 0x0006, type: 'in' },
        { id: 0x0008, type: 'in' },
        { id: 0x1000, type: 'in' },
        { id: 0x0019, type: 'out' }
      ]
    },
    {
      id: 0xf2,
      clusters: [
        { id: 0x0021, type: 'in' },
        { id: 0x0021, type: 'out' },
      ]
    }
  ]
});

class HueDimmerSwitch extends zigbee.Controller {
  constructor() {
    super();
    global.controllers.push(this);

    console.info("HueDimmerSwitch", this.device);

    for (const ep of this.device.endpoints) {
      this.device.read(0x0000, 0x0004, ep.endpoint);
      this.device.read(0x0000, 0x0005, ep.endpoint);
    }
  }

  onLeave() {
  }

  onUpdate() {

  }
}

HueDimmerSwitch.match = (matcher) => matcher.match({
  manufacturer: 'Philips',
  model: 'RWL020',
  profile: 0xc05e,
  endpoints: [
    {
      id: 0x01,
      clusters: [
        { id: 0x0000, type: 'in' },
        { id: 0x0000, type: 'out' },
        { id: 0x0003, type: 'out' },
        { id: 0x0004, type: 'out' },
        { id: 0x0005, type: 'out' },
        { id: 0x0006, type: 'out' },
        { id: 0x0008, type: 'out' }
      ]
    },
    {
      id: 0x02,
      clusters: [
        { id: 0x0000, type: 'in' },
        { id: 0x0001, type: 'in' },
        { id: 0x0003, type: 'in' },
        { id: 0x000f, type: 'in' },
        { id: 0xfc00, type: 'in' },
        { id: 0x0019, type: 'out' }
      ]
    }
  ]
});

const controllers = [
  HueAmbience,
  HueDimmerSwitch
];

zigbee.subscribe((matcher) => {
  const controller = controllers.find(controller => controller.match(matcher));

  if (controller != null) {
    matcher.subscribe(controller);
  }
});
