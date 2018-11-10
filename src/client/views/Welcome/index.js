/**
 * Sample React Native App
 * https://github.com/facebook/react-native
 *
 * @format
 * @flow
 */

import React from 'react';
import {Animated, Platform, StyleSheet, Text, View, Image} from 'react-native';
import { Pages, Indicator } from 'react-native-pages';
import { SharedElement } from 'react-native-motion';
import PropTypes from 'prop-types';
import Button from '../../components/Button';
import House from '../../components/House';
import Logo from '../../components/Logo';
import { constants } from '../../../../app.json';

type Props = {};
export default class Welcome extends React.Component {
  static contextTypes = {
    logoContainer: PropTypes.object.isRequired
  };

  constructor() {
    super();
    this.progress = new Animated.Value(0);

    this.translate = this.progress.interpolate({
      inputRange: [0, 0.6],
      outputRange: [0, -90],
      extrapolateLeft: 'clamp',
      extrapolateRight: 'clamp'
    });

    this.opacity = this.progress.interpolate({
      inputRange: [0, 0.5],
      outputRange: [1, 0],
      extrapolateLeft: 'clamp',
      extrapolateRight: 'clamp'
    });

    this.progress.addListener(this.onProgress);
  }

  componentDidMount() {
    this.timeout = setTimeout(this.onTimeout, 8000);
    this.onProgress();
  }

  componentWillUnmount() {
    clearTimeout(this.timeout);
  }

  onProgress = (value) => {
    const { logoContainer } = this.context;
    logoContainer.houseOpacity.setValue(1);
    logoContainer.translateY.setValue(this.translate.__getValue());
    logoContainer.tagLineOpacity.setValue(this.opacity.__getValue());
  };

  onTimeout = () => {
    if (this.refs.pages.scrollState !== -1) {
      return;
    }

    let progress = Math.floor(this.refs.pages.progress + 1);
    if(progress >= 4) {
      progress = 0;
    }

    this.refs.pages.scrollToPage(progress);
    this.timeout = setTimeout(this.onTimeout, 8000);
  };

  onScrollEnd = () => {
    if(this.timeout) {
      clearTimeout(this.timeout);
    }

    this.timeout = setTimeout(this.onTimeout, 8000);
  };

  renderPager = ({ horizontal, rtl, ...pager }) => {
    let { indicatorPosition } = pager;

    if ('none' === indicatorPosition) {
      return null;
    }

    let indicatorStyle = (horizontal && rtl)?
      pagerStyles.rtl:
      null;

    return (
      <View style={[pagerStyles[indicatorPosition], indicatorStyle, {zIndex: -1}]}>
        <Indicator {...pager} indicatorBorderColor={constants.foregroundColor} indicatorBorderWidth={1} />
      </View>
    );
  }

  onPressGetStarted = () => {
    const { history } = this.props;

    history.replace('/get-started');
  };

  render() {
    return (
      <View style={styles.container}>
        <View style={{ flex: 1 }}>
          <Pages
            ref="pages"
            onScrollEnd={this.onScrollEnd}
            progress={this.progress}
            indicatorColor={constants.foregroundColor}
            indicatorOpacity={0}
            renderPager={this.renderPager}
          >
            <View style={styles.blurbContainer}/>
            <View style={styles.blurbContainer}>
              <Text style={styles.blurb}>
                Do more with your home by using <Logo medium />. When you are needing to build
                your home automation, <Logo medium /> provides everything you need to make all
                your devices work together.
              </Text>
            </View>
            <View style={styles.blurbContainer}>
              <Text style={styles.blurb}>
                With <Logo medium /> everything in your home is simple. All devices will be
                connected to a central hub and nothing is too complex to be
                supported. <Logo medium /> provides for everything you want.
              </Text>
            </View>
            <View style={styles.blurbContainer}>
              <Text style={styles.blurb}>
                Everything you want to do in your home can be supported. <Logo medium /> makes
                your home easier to use.
              </Text>
            </View>
          </Pages>
        </View>
        <Button underlayColor={constants.highlightColor} style={styles.button} onPress={this.onPressGetStarted}>
          GET STARTED
        </Button>
      </View>
    );
  }
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    alignItems: 'stretch'
  },

  logoContainer: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
  },
  blurbContainer: {
    flex: 1,
    justifyContent: 'flex-end',
    alignItems: 'center',
    padding: 30,
    paddingBottom: 90
  },
  blurb: {
    fontFamily: constants.fontFamily,
    fontSize: 24,
    lineHeight: 37,
    color: constants.textColor,
    textAlign: 'center',
    maxWidth: 290
    //marginHorizontal: 30
  },
  button: {
    height: 75
  }
});

const pagerStyles = StyleSheet.create({
  rtl: {
    transform: [{
      rotate: '180deg',
    }],
  },

  container: {
    flex: 1,
  },

  bottom: {
    position: 'absolute',
    right: 0,
    bottom: 30,
    left: 0,
  },

  top: {
    position: 'absolute',
    top: 30,
    right: 0,
    left: 0,
  },

  left: {
    position: 'absolute',
    top: 0,
    bottom: 0,
    left: 20,
  },

  right: {
    position: 'absolute',
    top: 0,
    right: 20,
    bottom: 0,
  },

  indicator: {
    borderStyle: 'solid',
    borderColor: constants.foregroundColor,
    borderWidth: 1,
    backgroundColor: null
  }
});
