import React from 'react';
import {
  Animated,
  Image,
  StyleSheet,
  Text,
  View
} from 'react-native';
import { Switch } from 'react-router-native';
import PropTypes from 'prop-types';
import House from '../House';
import LogoImage from './logo.png';
import LogoMediumImage from './logo-medium.png';
import LogoSmallImage from './logo-small.png';
import { constants } from '../../../../app.json';

const styles = StyleSheet.create({
  container: {
    justifyContent: 'center',
    alignItems: 'center',
  },
  tagLine: {
    marginTop: 8,
    fontSize: 16,
    fontFamily: 'Lato',
    fontWeight: '400',
    color: '#4986A1'
  }
});

const containerStyles = StyleSheet.create({
  container: {
    flex: 1,
    alignItems: 'stretch'
  },

  logoContainer: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
  }
});


class Logo extends React.Component {
  render() {
    const { style, small, medium, tagLine, tagLineOpacity } = this.props;

    if (small) {
      return <Image source={LogoSmallImage} style={styles.small} />
    }

    if (medium) {
      return <Image source={LogoMediumImage} style={styles.medium} />
    }

    return (
      <Animated.View style={[styles.container, style]}>
        <Image source={LogoImage} />
        <Animated.View style={{opacity: tagLineOpacity}}>
          {!!tagLine && <Text style={styles.tagLine}>{constants.tagLine}</Text>}
        </Animated.View>
      </Animated.View>
    );
  }
}

Logo.Container = class LogoContainer extends React.Component {
  static childContextTypes = {
    logoContainer: PropTypes.object
  };

  constructor() {
    super();
    this.houseOpacity = new Animated.Value(0);
    this.tagLineOpacity = new Animated.Value(0);
    this.translateX = new Animated.Value(0);
    this.translateY = new Animated.Value(0);
    this.scale = new Animated.Value(1);
    this.childOpacity = new Animated.Value(0);
    this.nextChildrenOpacity = this.childOpacity.interpolate({
      inputRange: [0.5, 1],
      outputRange: [0, 1],
      extrapolateLeft: 'clamp',
      extrapolateRight: 'clamp'
    });
    this.previousChildrenOpacity = this.childOpacity.interpolate({
      inputRange: [0, 0.5],
      outputRange: [1, 0],
      extrapolateLeft: 'clamp',
      extrapolateRight: 'clamp'
    });

    this.state = {
      previousChildren: null,
      previousLocation: null
    };
  }

  componentDidMount() {
    Animated.timing(this.childOpacity, {
      toValue: 1,
      useNativeDriver: true
    }).start();
  }

  componentWillReceiveProps(nextProps) {
    const { childOpacity: animation } = this;

    if (!nextProps.location || !nextProps.location.animated) {
      animation.setValue(1);
      return;
    }

    this.setState((state) => {
      state.previousChildren = this.props.children;
      state.previousLocation = this.props.location;

      animation.setValue(0);
      Animated.timing(animation, {
        toValue: 1,
        useNativeDriver: true
      }).start(() => {
        this.setState({ previousChildren: null, previousLocation: null });
      });
    });
  }

  getChildContext() {
    return {
      logoContainer: this
    };
  }

  render() {
    const { previousChildren, previousLocation } = this.state;
    const { children: nextChildren, location: nextLocation } = this.props;

    const nextState = nextLocation.state || {};
    const { animated = true } = nextState;

    return (
      <View style={containerStyles.container}>
        <View style={{ flex: 1, position: 'relative'}}>
          <View pointerEvents="none" style={[StyleSheet.absoluteFill, {paddingBottom: 250}]}>
            <View style={[containerStyles.logoContainer]}>
              <House style={{ opacity: this.houseOpacity }}/>
            </View>
          </View>
          <View style={StyleSheet.absoluteFill}>
            {previousChildren && animated && <Animated.View key={previousLocation.key} style={[StyleSheet.absoluteFill, { opacity: this.previousChildrenOpacity }]}>
              <Switch location={previousLocation}>
                {previousChildren}
              </Switch>
            </Animated.View>}
            <Animated.View key={nextLocation.key} style={[StyleSheet.absoluteFill, { opacity: this.nextChildrenOpacity }]}>
              <Switch location={nextLocation}>
                {nextChildren}
              </Switch>
            </Animated.View>
          </View>
          <View pointerEvents="none" style={[StyleSheet.absoluteFill, {paddingBottom: 250}]}>
            <View style={[containerStyles.logoContainer]}>
              <Logo tagLine tagLineOpacity={this.tagLineOpacity} style={{ transform: [{ translateX: this.translateX }, { translateY: this.translateY }, { scale: this.scale }] }} />
            </View>
          </View>
        </View>
      </View>
    );
  }
}

export default Logo;
