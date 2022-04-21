class Handler:
    def run(self, req):
        return 'Method: %s\nBody: %s' % (req.method, str(req.json))
