// middleware.ts

import { NextResponse, NextRequest } from 'next/server';

// Trigger this middleware to run on the `/secret-page` route
export const config = {
    matcher: ['/edge-function'],
};

export function middleware(req: NextRequest) {
    console.log(`Visitor from ${req.geo?.country}`);

    // Rewrite to URL
    return new NextResponse('test edge function response');
    // return NextResponse.rewrite('https://www.polywork.com/mooseburger');
}
