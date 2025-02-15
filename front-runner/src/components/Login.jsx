import { Button, Form, Card } from 'react-bootstrap';
import './Login.css'; // Import custom styles

function Login() {
  return (
    <div className="login-container">
      <div className='logo-header'>
        <img src="../assets/Logo.svg" className="logo" alt="FR logo"/>
        <h1>FrontRunner</h1>
      </div>
      <Card className="login-card">
        <Card.Body style={{width:"100%"}}>
          <h2 className="text-center mb-4">Login</h2>
          <Form>
            <Form.Group className="mb-3" controlId="formBasicEmail">
              <Form.Label>Email address</Form.Label>
              <Form.Control type="email" placeholder="Enter email" />
            </Form.Group>

            <Form.Group className="mb-3" controlId="formBasicPassword">
              <Form.Label>Password</Form.Label>
              <Form.Control type="password" placeholder="Password" />
            </Form.Group>

            <Form.Group className="mb-3" controlId="formBasicCheckbox">
              <Form.Check type="checkbox" label="Remember me" />
            </Form.Group>

            <Button variant="primary" type="submit" className="login-button">
              Submit
            </Button>
          </Form>
        </Card.Body>
        <a href='/register'>New here? Create an account.</a>
      </Card>
      
    </div>
  );
}

export default Login;