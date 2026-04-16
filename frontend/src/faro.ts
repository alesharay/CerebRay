import { initializeFaro, getWebInstrumentations } from '@grafana/faro-react';  
import { TracingInstrumentation } from '@grafana/faro-web-tracing';  

initializeFaro({  
  url: 'https://faro-collector-prod-us-central-0.grafana.net/collect/0aedb11d9e4e5952c403b45e9dcc2d5c',  
  app: {  
    name: 'Cereb-Ray',  
    version: '0.0.0',  
    environment: import.meta.env.MODE,  
  },  
  instrumentations: [...getWebInstrumentations(), new TracingInstrumentation()],  
  ignoreErrors: [  
    /^ResizeObserver loop limit exceeded$/,  
    /^ResizeObserver loop completed with undelivered notifications$/,  
    /^Script error\.$/,  
    /chrome-extension:\/\//,  
    /moz-extension:\/\//,  
  ],  
});  
