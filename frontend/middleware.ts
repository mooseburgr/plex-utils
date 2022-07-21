// middleware.ts

import { NextResponse } from 'next/server'
import type { NextRequest } from 'next/server'

// Trigger this middleware to run on the `/secret-page` route
export const config = {
    matcher: ['/edge-function'],
};

export function middleware(req: NextRequest) {
    console.log(`Visitor from ${req.geo?.country}`);

    // Rewrite to URL
    // return new NextResponse('test edge function response');
    return NextResponse.redirect('https://www.polywork.com/mooseburger');
}
