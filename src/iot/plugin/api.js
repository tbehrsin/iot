
router.get('/hello-world', (req, res, next) => {
  setTimeout(() => {
    res.send(new B().message);
  }, 1000);
});

router.get('/async-hello-world', async (req, res, next) => {
  const test = await new Promise((resolve) => setTimeout(() => resolve(new B().message), 1000));
  res.send(test);
});


router.get('/fetch', async (req, res, next) => {
  try {
  const response = await fetch('http://www.google.com');
  const html = await response.text();
  res.send(html);
} catch(err) {
  console.error(err);
}
});

class A {
  constructor() {
    this.message = 'Hello World';
  }
}

class B extends A {

}
