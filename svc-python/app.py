import sys
from flask import Flask, escape, request, jsonify
sys.path.insert(1, 'lib')
from handler import Handler

# Initialize Flask
app = Flask(__name__)

@app.route('/',methods = ['POST', 'GET'])
def index():
#    content = request.json
#    print(content)
    return Handler().run(request)

if __name__ == '__main__':
    app.run(host="0.0.0.0", port=5000, debug=True)
