import React from 'react';

import {
  Animated,
  Image,
  StyleSheet,
  Text,
  TouchableWithoutFeedback,
  TouchableHighlight,
  View,
  ScrollView,
  Dimensions
} from 'react-native';
import { connect } from 'react-redux';
import PropTypes from 'prop-types';
import { Route, Switch, Link } from 'react-router-native';
import { SharedElement } from 'react-native-motion';
import { history } from '../../router';
import { constants } from '../../../../app.json';
import Button from '../../components/Button';
import Logo from '../../components/Logo';
import PinPad from '../../components/PinPad';
import Zone, { zones } from '../Zone';
import Device from '../Device';
import Settings from '../Settings';
import HouseImage from './house.png';
import BackArrowImage from './back-arrow.png';
import CogImage from './cog.png';
import StatusBarBackgroundImage from './status-bar-background.png';
import NavigationBackgroundImage from './navigation-background.png';

import * as actions from '../../actions';
import * as selectors from '../../selectors';
import { index as styles } from './styles';

const dim = Dimensions.get('window');

class Home extends React.Component {
  static contextTypes = {
    logoContainer: PropTypes.object.isRequired,
    router: PropTypes.object.isRequired
  };

  constructor() {
    super();
    this.state = {

    };

    this.scrollY = new Animated.Value(0);
    this.headerTranslateY = this.scrollY.interpolate({
      inputRange: [0, 176 + 48],
      outputRange: [0, -176 - 48],
      extrapolateRight: 'clamp'
    })
  }

  componentDidMount() {
    const { logoContainer } = this.context;

    Animated.timing(logoContainer.houseOpacity, { toValue: 0 }).start();
    Animated.timing(logoContainer.tagLineOpacity, { toValue: 0 }).start();
    Animated.timing(logoContainer.translateY, { toValue: -(dim.height - 250) / 2 + 54 * 0.6111 + 23 }).start();
    Animated.timing(logoContainer.translateX, { toValue: dim.width / 2 - 77 * 0.6111 / 2 - 16 }).start();
    Animated.timing(logoContainer.scale, { toValue: 0.6111 }).start();
  }

  onPressBack = () => {
    history.goBack();
  };

  renderZone = (zone) => {
    return (
      <View key={zone.id} style={styles.zoneOn}>
        <Link to={{ pathname: `/zones/${zone.id}`, state: { animated: false, title: zone.name } }} activeOpacity={0.8} underlayColor={constants.backgroundColor}>
          <View style={styles.zoneWrapper}>
            <Text style={styles.zoneTitle}>{zone.name}</Text>
          </View>
        </Link>
      </View>
    );
  }

  render() {
    return (
      <View style={styles.container}>
        <Animated.ScrollView
          contentContainerStyle={[styles.scrollContentContainer, {paddingTop: history.index === 0 ? 227 : 80}]}
          showsVerticalScrollIndicator={false}
          scrollEnabled={!history.location.state || history.location.state.scrollEnabled !== false}
          onScroll={Animated.event(
            [{
              nativeEvent: { contentOffset: { y: this.scrollY }}
            }],
            { useNativeDriver: true }
          )}
          scrollEventThrottle={16}
        >
          <Switch>
            <Route exact path="/devices/:id" component={Device} />
            <Route exact path="/zones/:id" component={Zone} />
            <Route path="/settings/" component={Settings} />
            <Route exact path="/" render={(props) => (<View style={styles.zoneContainer}>{zones.map(zone => this.renderZone(zone))}</View>)} />
          </Switch>
        </Animated.ScrollView>
        <View pointerEvents="box-none" style={styles.header}>
          <View pointerEvents="none" tyle={styles.navigationBackground}>
            <Image source={NavigationBackgroundImage} style={styles.navigationBackgroundImage} />
          </View>
          <View style={styles.navigation}>
            {history.index !== 0 && (
              <TouchableHighlight onPressIn={this.onPressBack} activeOpacity={0.8} underlayColor={constants.backgroundColor}>
                <Image source={BackArrowImage} />
              </TouchableHighlight>
            )}
            {history.index === 0 && (
              <Link to={{ pathname: `/settings`, state: { animated: false, title: 'Settings' } }} activeOpacity={0.8} underlayColor={constants.backgroundColor}>
                <Image source={CogImage} />
              </Link>
            )}
            {history.location.state && history.location.state.title && (
              <View style={styles.navigationTitle}>
                <Text style={styles.navigationTitleText}>{history.location.state.title}</Text>
              </View>
            )}
          </View>
          {history.index === 0 && (
            <Animated.View pointerEvents="none" style={[styles.house, { transform: [{ translateY: this.headerTranslateY }]}]}>
              <Image source={HouseImage} />
            </Animated.View>
          )}
          <View pointerEvents="none" style={styles.statusBarBackground}>
            <Image source={StatusBarBackgroundImage} style={styles.statusBarBackgroundImage} />
          </View>
        </View>
      </View>
    );
  }
}


const mapDispatchToProps = (dispatch, props) => ({
  request: (key, options) => dispatch(actions.api.request(key, options)),
  ...props
});
const mapStateToProps = (state, props) => ({
  ...props
});
export default connect(mapStateToProps, mapDispatchToProps)(Home);
