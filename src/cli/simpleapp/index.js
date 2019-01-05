
class SimpleController extends zigbee.Controller {
  constructor() {
    super();
    this.name = `Test Device ${this.device.eui64}`;
    this.setState({
      on: false,
      level: 254,
      xy: [0, 0]
    }, true);
    this.setStateRecursive().catch(error => console.error(error));
  }

  async setStateRecursive() {
    this.timeout = setTimeout(() => this.pollAttributes().catch(error => console.error(error)), 1000);

    this.setState({ on: !this.state.on }, true);
    this.setState({ level: (this.state.level + 1) % 254 }, true);

    let theta = Math.atan2(this.xy[1], this.xy[0]);
    theta += Math.PI * (30 / 360);
    this.setState({ xy: [Math.cos(theta), Math.sin(theta)] }, true);
  }

  onLeave() {

  }

  onUpdate() {

  }

  onSetState(state) {
    return state
  }
}

let subscribed = false;
zigbee.subscribe((matcher) => {
  if (!subscribed) {
    subscribed = true;
    matcher.subscribe(SimpleController);
  }
});
