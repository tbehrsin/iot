import {
  Dimensions,
  StyleSheet
} from 'react-native';
import { constants } from '../../../../app.json';

const dim = Dimensions.get('window');
const width = dim.width / 2 - 30;

export const index = StyleSheet.create({
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
  scrollContentContainer: {
    flexGrow: 1
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
  zoneContainer: {
    flexDirection: 'row',
    flexWrap: 'wrap',
    justifyContent: 'center',
    paddingBottom: 10
  },
  zoneOn: {
    margin: 10
  },
  zoneTitle: {
    fontFamily: constants.fontFamily,
    fontSize: constants.subtitleFontSize,
    fontWeight: constants.subtitleFontWeight,
    color: constants.foregroundColor
  },
  zoneWrapper: {
    width,
    height: width * 140 / 160,
    padding: 8,
    borderColor: constants.foregroundColor,
    backgroundColor: constants.boxOnColor,
    borderWidth: 1,
    borderStyle: 'solid',
  },
  navigationTitle: {
    marginLeft: 18
  },
  navigationTitleText: {
    fontFamily: constants.fontFamily,
    fontSize: constants.titleFontSize,
    fontWeight: constants.titleFontWeight,
    color: constants.foregroundColor
  }
});
