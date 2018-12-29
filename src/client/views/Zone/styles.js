import {
  StyleSheet,
  Dimensions
} from 'react-native';
import { constants } from '../../../../app.json';

const dim = Dimensions.get('window');
const width = dim.width / 2 - 30;

export const index = StyleSheet.create({
  container: {
    flex: 1,
    flexDirection: 'row',
    flexWrap: 'wrap',
    justifyContent: 'flex-start',
    paddingLeft: 6
  },
  box: {
    margin: 12
  },
  boxTitle: {
    fontFamily: constants.fontFamily,
    fontSize: constants.subtitleFontSize,
    fontWeight: constants.subtitleFontWeight,
    color: constants.textColor
  },
  boxOnTitle: {
    color: constants.foregroundColor
  },
  boxWrapper: {
    width,
    height: width * 140 / 160,
    padding: 8,
    borderColor: constants.textColor,
    backgroundColor: constants.boxOffColor,
    borderWidth: 1,
    borderStyle: 'solid',
  },
  boxOnWrapper: {
    backgroundColor: constants.boxOnColor,
    borderColor: constants.foregroundColor,
  },
  boxSpacer: {
    flex: 1
  },
  boxButton: {
    backgroundColor: constants.textColor,
    marginHorizontal: -8,
    marginBottom: -8,
    padding: 8
  },
  boxOnButton: {
    backgroundColor: constants.foregroundColor
  }
});
