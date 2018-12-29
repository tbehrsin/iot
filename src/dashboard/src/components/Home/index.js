
import React from 'react';
import { Router, Route, Redirect } from 'react-router-dom';
import { Provider } from 'react-redux';
import { PersistGate } from 'redux-persist/integration/react';
import { store, persistor } from '../../store';
import { history } from '../../router';
import Logo from '../Logo';
import houseImage from './house@2x.png';

import Page from '../Page';

import styles from './index.scss';

class Home extends React.Component {

  renderHeader = () => {
    return (
      <div className={styles.header}>
        <div className={styles.text}>
          Do more with your home using <Logo />. When you need to build your home automation, <Logo /> provides
          everything you need to make all your home automation devices work together.
        </div>
        <img src={houseImage} />
      </div>
    );
  };

  render() {
    return (
      <Page className={styles.content} header={this.renderHeader} />
    );
  }
}

export default Home;
