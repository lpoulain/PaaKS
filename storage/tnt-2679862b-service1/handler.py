class Handler:
    def run(self, req):
        return 'METHOD: %s\nBODY: %s' % (req.method, str(req.json))
