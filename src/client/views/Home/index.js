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
import PropTypes from 'prop-types';
import { SharedElement } from 'react-native-motion';
import Button from '../../components/Button';
import Logo from '../../components/Logo';
import PinPad from '../../components/PinPad';
import { constants } from '../../../../app.json';
import HouseImage from './house.png';
import BackArrowImage from './back-arrow.png';
import StatusBarBackgroundImage from './status-bar-background.png';
import NavigationBackgroundImage from './navigation-background.png';

const styles = StyleSheet.create({
  container: {
    flex: 1,
    alignItems: 'stretch'
  },
  header: {
    position: 'absolute',
    top: 0,
    left: 0,
    right: 0
  },
  navigationBackground: {
    position: 'absolute'
  },
  statusBarBackground: {
    position: 'absolute'
  },
  navigation: {
    position: 'absolute',
    marginTop: 20,
    flexDirection: 'row',
    height: 60,
    alignItems: 'center',
    paddingHorizontal: 16
  },
  house: {
    top: -34,
    alignSelf: 'center'
  },
  boxOn: {
    borderColor: constants.foregroundColor,
    backgroundColor: constants.boxOnColor,
    borderWidth: 1,
    borderStyle: 'solid',

  }
});

const dim = Dimensions.get('window');

class Home extends React.Component {
  static contextTypes = {
    logoContainer: PropTypes.object.isRequired
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

  renderBox = (_, i) => {
    return (
      <View key={i} style={styles.boxOn}>
        <Text>{i}</Text>
      </View>
    );
  }

  render() {

    const boxes = Array(100).fill().map(this.renderBox);

    return (
      <View style={styles.container}>
        <Animated.ScrollView
          contentContainerStyle={{paddingTop: 227}}
          showsVerticalScrollIndicator={false}
          onScroll={Animated.event(
            [{
              nativeEvent: { contentOffset: { y: this.scrollY }}
            }],
            { useNativeDriver: true }
          )}
          scrollEventThrottle={16}
        >
          {boxes}
        </Animated.ScrollView>
        <View pointerEvents="none" style={styles.header}>
          <View pointerEvents="none" tyle={styles.navigationBackground}>
            <Image source={NavigationBackgroundImage} style={styles.navigationBackgroundImage} />
          </View>
          <View style={styles.navigation}>
            <Image source={BackArrowImage} />
          </View>
          <Animated.View pointerEvents="none" style={[styles.house, { transform: [{ translateY: this.headerTranslateY }]}]}>
            <Image source={HouseImage} />
          </Animated.View>
          <View pointerEvents="none" style={styles.statusBarBackground}>
            <Image source={StatusBarBackgroundImage} style={styles.statusBarBackgroundImage} />
          </View>
        </View>
      </View>
    );
  }
}

export default Home;
