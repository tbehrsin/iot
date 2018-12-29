
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

export const zones = [
  { id: '0', name: 'Kitchen' },
  { id: '1', name: 'Living Room' },
  { id: '2', name: 'Master Bedroom' },
  { id: '3', name: 'Ken\'s Bedroom' },
  { id: '4', name: 'Ben\'s Bedroom' },
  { id: '5', name: 'Len\'s Bedroom' },
  { id: '6', name: 'Dining Room' },
  { id: '7', name: 'Utility Room' },
];


class Zone extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
    };
  }

  componentDidMount() {
    const { request } = this.props;

    this.stateSubscription = ws.subscribe('iot/+/notify', this.onNotify);
    request('devices', { path: '/api/v1/devices/' });
  }

  componentWillUnmount() {
    this.stateSubscription.unsubscribe();
  }

  onNotify = (topic, device, [, id]) => {
    const { devices, setIn } = this.props;

    const index = devices.findIndex(device => id === device.id);
    if (index !== -1) {
      setIn('devices', [index], device);
    }
  };

  onClickOn = (device, event) => {
    const { setOn } = this.props;

    ws.publish(`iot/${device.id}/publish`, { state: { on: !device.state.on } });
    event.preventDefault();
  };

  renderDevice(device) {
    const { state } = device;

    return (
      <Link
        to={{ pathname: `/dashboard/devices/${device.id}`, state: { title: device.name, scrollEnabled: false } }}
        key={device.id}
        className={`${styles.box} ${state && state.on ? styles.boxOn : ''}`}
      >
        <div className={styles.boxTitle}>{device.name}</div>
        <div className={styles.spacer} />
        {state && 'on' in state && (
          <Button className={styles.boxButton} onClick={(event) => this.onClickOn(device, event)}>{state.on ? 'ON' : 'OFF'}</Button>
        )}
      </Link>
    );
  }

  render() {
    const { zone, devices } = this.props;

    return (
      <Page.FullScreen className={styles.container}>
        {devices && devices.map(device => this.renderDevice(device))}
      </Page.FullScreen>
    );
  }
}

const mapDispatchToProps = (dispatch, props) => ({
  reset: (key) => dispatch(actions.api.reset(key)),
  request: (key, options) => dispatch(actions.api.request(key, options)),
  setIn: (key, path, value) => dispatch(actions.api.setIn(key, path, value)),
  ...props
});
const mapStateToProps = (state, props) => ({
  zone: zones[props.match.params.id],
  devices: selectors.api.body(state)('devices'),
  ...props
});
export default connect(mapStateToProps, mapDispatchToProps)(Zone);
