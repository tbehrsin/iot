import React from 'react';

import {
  Animated,
  Image,
  StyleSheet,
  Text,
  TouchableWithoutFeedback,
  View,
  ScrollView,
  Dimensions
} from 'react-native';
import { connect } from 'react-redux';
import PropTypes from 'prop-types';
import { Route, Switch, Link } from 'react-router-native';

import { ws } from '../../services';
import Button from '../../components/Button';

import * as actions from '../../actions';
import * as selectors from '../../selectors';
import { index as styles } from './styles';

const dim = Dimensions.get('window');

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

    const { reset } = props;
    reset('list-devices');
  }

  componentDidMount() {
    const { request } = this.props;

    this.stateSubscription = ws.subscribe('iot/+/state', this.onSetState);

    request('list-devices', { path: '/api/v1/devices/' });
  }

  componentWillUnmount() {
    this.stateSubscription.unsubscribe();
  }

  onSetState = (topic, state) => {
    const { devices, setIn } = this.props;

    const [,id] = topic.match(/^iot\/([^/]+)\/state$/);

    console.debug(topic, state);

    const index = devices.findIndex(device => id === device.id);
    if (index !== -1) {
      setIn('list-devices', [index, 'state'], state);
    }
  };

  onPressOn = (device) => {
    const { setOn } = this.props;

    setOn(device.id, !device.state.on);
  };

  renderDevice(device) {
    const { state } = device;

    return (
      <View key={device.id} style={styles.box}>
        <Link to={{ pathname: `/devices/${device.id}`, state: { title: device.name, scrollEnabled: false } }}>
          <View style={[styles.boxWrapper, state.on ? styles.boxOnWrapper : null]}>
            <Text style={[styles.boxTitle, state.on ? styles.boxOnTitle : null]}>{device.name}</Text>
            <View style={styles.boxSpacer} />
            <Button style={[styles.boxButton, state.on ? styles.boxOnButton : null]} onPress={() => this.onPressOn(device)}>{state.on ? 'ON' : 'OFF'}</Button>
          </View>
        </Link>
      </View>
    );
  }

  render() {
    const { zone, devices } = this.props;

    return (
      <View style={styles.container}>
        {devices && devices.map(device => this.renderDevice(device))}
      </View>
    );
  }
}


const mapDispatchToProps = (dispatch, props) => ({
  reset: (key) => dispatch(actions.api.reset(key)),
  request: (key, options) => dispatch(actions.api.request(key, options)),
  setIn: (key, path, value) => dispatch(actions.api.setIn(key, path, value)),
  setOn: (id, on) => dispatch(actions.api.request(null, { method: 'PATCH', path: `/api/v1/devices/${id}/`, body: { state: { on } } })),
  ...props
});
const mapStateToProps = (state, props) => ({
  zone: zones[props.match.params.id],
  devices: selectors.api.body(state)('list-devices'),
  ...props
});
export default connect(mapStateToProps, mapDispatchToProps)(Zone);
