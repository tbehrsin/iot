import React from 'react';
import {
  Animated,
  View,
  Image,
  StyleSheet,
  Dimensions
} from 'react-native';
import HouseImage from './house.png';

const dim = Dimensions.get('window');

const styles = StyleSheet.create({
  default: {
    position: 'absolute',
    width: dim.width,
    alignItems: 'center'
  }
})

class House extends React.Component {
  render() {
    const { style } = this.props;

    return (
      <Animated.View style={[styles.default, style]}>
        <Image source={HouseImage} />
      </Animated.View>
    );
  }
}

export default House;
