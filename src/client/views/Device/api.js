import { constants as config } from '../../../../app.json';

const css = `
html, body {
  margin: 0;
  padding: 0;
  font-family: 'lato', sans-serif;
  font-weight: 400;
  font-size: 18px;
  color: ${config.textColor};
  -webkit-font-smoothing: subpixel-antialiased;
}

body {
  padding: 0 12px 12px;
}
`;


export default `
const meta = document.createElement('meta');
meta.name = 'viewport';
meta.content = 'width=device-width, initial-scale=1.0, user-scalable=no';

const style = document.createElement('style');
style.innerHTML = \`${css}\`;

document.head.append(meta, style);
`;
