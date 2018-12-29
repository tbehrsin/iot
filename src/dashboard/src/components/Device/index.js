
import React from 'react';
import { connect } from 'react-redux';
import { Router, Route, Redirect, Link } from 'react-router-dom';
import { Provider } from 'react-redux';
import { PersistGate } from 'redux-persist/integration/react';
import { store, persistor } from '../../store';
import { history } from '../../router';

import * as actions from '../../actions';
import * as selectors from '../../selectors';
import { ws } from '../../services';

import Page from '../Page';
import Button from '../Button';

import styles from './index.scss';


class Device extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
    };
  }

  componentDidMount() {
    const { device, request, match } = this.props;

    this.stateSubscription = ws.subscribe(`iot/${match.params.id}/notify`, this.onNotify);

    if (!device) {
      request('devices', { path: '/api/v1/devices/' });
    }

    this.submitForm();
  }

  componentWillUnmount() {
    this.stateSubscription.unsubscribe();

    window.removeEventListener('message', this.onMessage, false);
  }

  componentDidUpdate() {
    this.submitForm();
  }

  componentWillReceiveProps(nextProps) {
    if (this.port && ('device' in nextProps)) {
      const { device } = nextProps;
      this.port.postMessage(JSON.stringify(device));
    }
  }

  onNotify = (topic, device) => {
    const { devices, setIn } = this.props;

    const index = devices.findIndex(d => d.id === device.id);
    if (index !== -1) {
      setIn('devices', [index], device);
    }
  }

  onLoad = () => {
    const { url, device } = this.props;
    const { iframe } = this.refs;

    const channel = new MessageChannel();
    this.port = channel.port1;
    this.port.onmessage = this.onMessage;

    iframe.contentWindow.postMessage('iot-frame-detector', '*', [channel.port2]);
    this.port.postMessage(JSON.stringify({ urlPrefix: url }));
    this.port.postMessage(JSON.stringify(device));
  };

  onMessage = (event) => {
    const { state } = JSON.parse(event.data);
    const { request, match } = this.props;

    ws.publish(`iot/${match.params.id}/publish`, { state });

    // request(`patch-device:${match.params.id}`, { path: `/api/v1/devices/${match.params.id}/`, method: 'PATCH', body: { state } });
  };

  shouldComponentUpdate() {
    const { device } = this.props;
    const { form, iframe } = this.refs;

    if (device && form) {
      if (!this.submitted) {
        this.submitForm();
      }
      return false;
    }

    return true;
  }

  submitForm() {
    const { device } = this.props;
    const { form } = this.refs;

    if (device && form) {
      this.submitted = true;
      form.submit();
    }
  }

  render() {
    const { device, url, authToken } = this.props;

    return (
      <Page.FullScreen className={styles.container}>
        {device && device.state && (
          <div className={styles.wrapper}>
            <form ref="form" action={`${url}${device.id}/public/`} method="POST" target="device">
              <input type="hidden" name="__authToken" value={authToken} />
            </form>
            <iframe ref="iframe" name="device" onLoad={this.onLoad} src="about:empty" />
          </div>
        )}
        {device && !device.state && (
          <div className={styles.unassigned}>
            <div className={styles.unassignedTitle}>
              <i className="fas fa-cloud-download-alt" />
              <span>not assigned to an app</span>
            </div>
            <Button component={Link} className={styles.assignButton} to={`/dashboard/devices/${device.id}/assign`}>Assign to app</Button>
          </div>
        )}
      </Page.FullScreen>
    );
  }
}

const mapDispatchToProps = (dispatch, props) => ({
  request: (key, options) => dispatch(actions.api.request(key, options)),
  setIn: (key, path, value) => dispatch(actions.api.setIn(key, path, value)),
  ...props
});
const mapStateToProps = (state, props) => ({
  devices: selectors.api.body(state)('devices'),
  device: selectors.api.findInBody(state)('devices', d => d.get('id') === props.match.params.id),
  authToken: selectors.auth.getToken(state),
  url: selectors.api.url(state)('devices'),
  ...props
});
export default connect(mapStateToProps, mapDispatchToProps)(Device);
