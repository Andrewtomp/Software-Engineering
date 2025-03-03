// NavBar.test.js
import { render, fireEvent,screen } from '@testing-library/react';
import NavBar from '../NavBar';

describe('NavBar', () => {
    it('navigates to the correct links when clicked', () => {
        // Mock the window.location.href to check navigation
        const mockLocation = { href: '' };
        delete window.location;
        window.location = mockLocation;

        render(<NavBar />);
        fireEvent.click(screen.getByText(/frontrunner/i));
        expect(window.location.href).toBe('/');
        
        fireEvent.click(screen.getByText(/my products/i));
        expect(window.location.href).toBe('/products');

        fireEvent.click(screen.getByText(/my storefronts/i));
        expect(window.location.href).toBe('/storefronts');

        fireEvent.click(screen.getByText(/my orders/i));
        expect(window.location.href).toBe('/orders');

    });
});
