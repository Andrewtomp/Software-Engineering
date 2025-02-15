import React from 'react';
import { Button, Form, Card } from 'react-bootstrap';
import './Registration.css'; // Import CSS for styling

function Registration() {
  return (
    <div className="register-container">
      <div className='logo-header'>
        <img src="../assets/Logo.svg" className="logo" alt="FR logo"/>
        <h1>FrontRunner</h1>
      </div>
      <Card className="register-card">
        <Card.Body style={{width:"100%"}}>
          <h2 className="text-center mb-4">Register</h2>
          <Form>
            <Form.Group className="mb-3" controlId="formBasicName">
              <Form.Label>Business Name</Form.Label>
              <Form.Control type="text" placeholder="Enter business name" />
            </Form.Group>

            <Form.Group className="mb-3" controlId="formBasicEmail">
              <Form.Label>Email address</Form.Label>
              <Form.Control type="email" placeholder="Enter email" />
              <Form.Text className="text-muted">
                We'll never share your email with anyone else.
              </Form.Text>
            </Form.Group>

            <Form.Group className="mb-3" controlId="formBasicPassword">
              <Form.Label>Password</Form.Label>
              <Form.Control type="password" placeholder="Password" />
            </Form.Group>

            <Form.Group className="mb-3" controlId="formConfirmPassword">
              <Form.Label>Confirm Password</Form.Label>
              <Form.Control type="password" placeholder="Confirm password" />
            </Form.Group>

            <Button variant="primary" type="submit" className="register-button">
              Submit
            </Button>
          </Form>
        </Card.Body>
        <a href='/login'>Already have an account? Log in.</a>
      </Card>
    </div>
  );
}

export default Registration;
