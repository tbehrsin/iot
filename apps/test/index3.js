

class Test extends zigbee.Test {


  test() {
    const data = new ArrayBuffer(8);
    const view = new DataView(data);
    view.setUint16(0, 0x1234, true);
    view.setUint16(2, 0x5678, true);
    view.setUint16(4, 0x9876, true);

    const array = new Uint8Array(data);
    console.info(array.toString());
  }

}

global.test = new Test();

global.Test = Test;
