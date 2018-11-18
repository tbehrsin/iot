
const router = new Router();
app.use('/s', router);

router.get('/async-hello-world', (req, res, next) => {
  res.json({
    hello: 'world'
  });
});


const router2 = new Router();

router.use('/2/', router2);

router2.get('/hello', (req, res, next) => {
  res.json({
    hello: 'world2'
  });
})

router.get('/async-hello-world', async (req, res, next) => {
  console.info("handler 2");
  //const test = await new Promise((resolve) => setTimeout(() => resolve('Hello World'), 1000));
  res.send(`${"Hello World"}\n`);
});

try {
  class HueGo extends zigbee.Device {
    constructor() {
      super();
      console.info('new HueGoController()', this.props.eui64);
      console.info(`props(${this.props.eui64}): ${JSON.stringify(this.props, null, 2)}`);
    }

    onLeave() {
      console.info('HueGoController#onLeave()', this.props.eui64);
      console.info(`props(${this.props.eui64}): ${JSON.stringify(this.props, null, 2)}`);
    }

    onUpdate() {
      console.info('HueGoController#onUpdate()', this.props.eui64);
      console.info(`props(${this.props.eui64}): ${JSON.stringify(this.props, null, 2)}`);
    }
  }

  zigbee.subscribe((device) => device.subscribe(HueGo));
} catch (error) {
  console.error(error);
}
